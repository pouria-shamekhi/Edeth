/* ============================================
   Edeth - Main Application
   ============================================ */

const AUDIO_EXTS = ['.mp3', '.wav', '.flac', '.aac', '.ogg', '.wma', '.m4a', '.opus', '.srm', '.pcm', '.raw'];
const VIDEO_EXTS = ['.mp4', '.mkv', '.avi', '.mov', '.webm', '.flv', '.wmv'];
const SUB_EXTS = ['.srt', '.ass', '.ssa', '.vtt', '.sub'];
const MEDIA_EXTS = [...AUDIO_EXTS, ...VIDEO_EXTS];

const App = {
  currentPage: 'home',

  init() {
    window.addEventListener('hashchange', () => this.route());
    this.setupNavigation();
    this.route();
  },

  setupNavigation() {
    document.querySelectorAll('.nav-link').forEach(link => {
      link.addEventListener('click', () => {
        document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
        link.classList.add('active');
      });
    });
  },

  route() {
    const hash = (window.location.hash || '#home').slice(1);
    this.currentPage = hash;

    document.querySelectorAll('.nav-link').forEach(link => {
      link.classList.toggle('active', link.dataset.page === hash);
    });

    const content = document.getElementById('content');
    content.innerHTML = '';

    const pages = {
      home:      () => this.renderHome(),
      watermark: () => this.renderWatermark(),
      subtitle:  () => this.renderSubtitle(),
      trim:      () => this.renderTrim(),
      extract:   () => this.renderExtract(),
      merge:     () => this.renderMerge(),
      compress:  () => this.renderCompress()
    };

    const renderFn = pages[hash] || pages.home;
    content.appendChild(renderFn());
  },

  // ==============================
  // HOME PAGE
  // ==============================
  renderHome() {
    const page = document.createElement('div');

    // Futuristic EDITH Hero Banner
    const hero = document.createElement('div');
    hero.className = 'edeth-hero-banner';
    hero.innerHTML = `
      <div class="hero-content">
        <div class="hero-logo-box">
          <img src="img/logo.png" alt="Edeth AI Eye" class="hero-logo-img">
        </div>
        <div class="hero-text-box">
          <div class="hero-badge">
            <span class="hero-badge-dot"></span>
            <span>EDETH MULTIMEDIA CORE 2.0</span>
          </div>
          <h1 class="hero-title">Next-Gen Audio DSP & Video Suite</h1>
          <p class="hero-desc">Advanced FFT watermark cancellation with phase inversion, zero-latency A/B comparison visualizer, smart constrained-bitrate video compression, and lossless FFmpeg pipeline.</p>
        </div>
      </div>
    `;
    page.appendChild(hero);

    const grid = document.createElement('div');
    grid.className = 'feature-grid';

    // Main feature: Watermark Remover
    grid.appendChild(EdethUI.createFeatureCard(
      'Audio Watermark Remover',
      'Remove repeating audio watermarks using FFT cross-correlation and least-squares phase cancellation. Provide the watermarked track and the raw watermark sample — Edeth removes only the watermark in seconds.',
      'wand', 'watermark', true
    ));

    grid.appendChild(EdethUI.createFeatureCard(
      'Add Subtitles',
      'Burn subtitles permanently into video or embed as a selectable text stream (SRT, ASS, VTT).',
      'subtitle', 'subtitle'
    ));

    grid.appendChild(EdethUI.createFeatureCard(
      'Trim Media',
      'Cut audio or video to exact timestamps using fast stream copy without quality loss.',
      'scissors', 'trim'
    ));

    grid.appendChild(EdethUI.createFeatureCard(
      'Extract Audio',
      'Separate the audio track from video files into MP3, WAV, AAC, or FLAC.',
      'music', 'extract'
    ));

    grid.appendChild(EdethUI.createFeatureCard(
      'Merge Audio',
      'Add or replace audio in video files with looping to match video duration or custom time offsets.',
      'merge', 'merge'
    ));

    grid.appendChild(EdethUI.createFeatureCard(
      'Compress Video',
      'Smart 2-pass & constrained-bitrate compressor for Discord (25MB), Telegram (50MB), or web optimization.',
      'compress', 'compress'
    ));

    page.appendChild(grid);
    return page;
  },

  // ==============================
  // WATERMARK REMOVER PAGE
  // ==============================
  renderWatermark() {
    const page = document.createElement('div');

    page.appendChild(EdethUI.createPageHeader(
      'Audio Watermark Remover',
      'Automatically find and erase audio watermarks using FFT cross-correlation & least-squares subtraction.'
    ));

    const form = document.createElement('div');

    const inputFile = EdethUI.createFileInput('wm-input', 'Watermarked Song / Audio Track', AUDIO_EXTS);
    form.appendChild(inputFile);

    const wmFile = EdethUI.createFileInput('wm-watermark', 'Raw Watermark Audio Sample (.mp3, .srm, .wav, etc.)', AUDIO_EXTS);
    form.appendChild(wmFile);

    const outputFile = EdethUI.createFileInput('wm-output', 'Output Clean Audio File (auto-generated if empty)', AUDIO_EXTS);
    form.appendChild(outputFile);

    // Audio preview controls for source tracks
    const previewRow = document.createElement('div');
    previewRow.style.display = 'flex';
    previewRow.style.gap = '10px';
    previewRow.style.margin = '10px 0 16px 0';
    previewRow.style.flexWrap = 'wrap';

    const previewInputBtn = EdethUI.createButton('🎵 Preview Input Track', () => {
      const p = inputFile.getValue();
      if (!p) { EdethUI.showToast('Please choose an input audio file first', 'info'); return; }
      const existing = form.querySelector('#wm-input-preview');
      if (existing) { existing.remove(); return; }
      const player = EdethUI.createAudioPlayer('wm-input-preview', p, 'Input Track Preview');
      form.insertBefore(player, previewRow.nextSibling);
    }, 'secondary', 'sm');

    const previewWmBtn = EdethUI.createButton('🔊 Preview Watermark Sample (.srm / .mp3)', () => {
      const p = wmFile.getValue();
      if (!p) { EdethUI.showToast('Please choose a watermark sample file first', 'info'); return; }
      const existing = form.querySelector('#wm-sample-preview');
      if (existing) { existing.remove(); return; }
      const player = EdethUI.createAudioPlayer('wm-sample-preview', p, 'Watermark Sample Preview');
      form.insertBefore(player, previewRow.nextSibling);
    }, 'secondary', 'sm');

    previewRow.appendChild(previewInputBtn);
    previewRow.appendChild(previewWmBtn);
    form.appendChild(previewRow);

    inputFile.getInput().addEventListener('change', () => {
      const val = inputFile.getValue();
      if (val && !outputFile.getValue()) {
        outputFile.setValue(EdethUI.generateOutputName(val, '_clean'));
      }
    });

    const progress = EdethUI.createProgressBar('wm-progress');
    const terminal = EdethUI.createTaskTerminal('wm-terminal');

    const submitBtn = EdethUI.createButton('Remove Watermark', async () => {
      const input = inputFile.getValue();
      const wm = wmFile.getValue();
      const output = outputFile.getValue() || EdethUI.generateOutputName(input, '_clean');

      if (!input) { EdethUI.showToast('Please select a watermarked audio file', 'error'); return; }
      if (!wm) { EdethUI.showToast('Please select a watermark sample file', 'error'); return; }

      submitBtn.disabled = true;
      outputFile.setValue(output);

      // Clear previous terminal logs
      const body = document.getElementById('wm-terminal-body');
      if (body) body.innerHTML = '';

      // Collect detected watermark timestamps from SSE logs
      const detectedTimestamps = [];

      try {
        const data = await this.apiCall('/api/watermark/remove', {
          inputFile: input,
          watermarkFile: wm,
          outputFile: output
        });

        // Listen for watermark timestamps
        const termEl = document.getElementById('wm-terminal');
        const observer = new MutationObserver(() => {
          if (!termEl) return;
          const lines = termEl.querySelectorAll('.terminal-line');
          lines.forEach(l => {
            const m = l.textContent.match(/at (\d+\.?\d*)s/);
            if (m) {
              const sec = parseFloat(m[1]);
              if (!detectedTimestamps.includes(sec)) {
                detectedTimestamps.push(sec);
              }
            }
          });
        });
        const termBody = document.getElementById('wm-terminal-body');
        if (termBody) observer.observe(termBody, { childList: true });

        this.subscribeProgress(data.taskId, 'wm-progress', 'wm-terminal', (res) => {
          submitBtn.disabled = false;
          observer.disconnect();

          if (res && res.status === 'complete') {
            const oldAB = page.querySelector('#wm-ab-player');
            if (oldAB) oldAB.remove();

            const abPlayer = EdethUI.createABComparisonPlayer('wm-ab-player', input, output, detectedTimestamps);
            page.appendChild(abPlayer);
          }
        });
      } catch (err) {
        submitBtn.disabled = false;
      }
    }, 'primary', 'lg');

    form.appendChild(document.createElement('br'));
    form.appendChild(submitBtn);
    form.appendChild(progress);
    form.appendChild(terminal);

    const card = EdethUI.createCard('Watermark Removal Pipeline', form);
    page.appendChild(card);
    return page;
  },

  // ==============================
  // SUBTITLE PAGE
  // ==============================
  renderSubtitle() {
    const page = document.createElement('div');

    page.appendChild(EdethUI.createPageHeader(
      'Add Subtitles',
      'Add subtitle tracks to videos — burn them permanently into video frames or embed as a selectable stream.'
    ));

    const form = document.createElement('div');

    const videoFile = EdethUI.createFileInput('sub-video', 'Video File', VIDEO_EXTS);
    form.appendChild(videoFile);

    const subFile = EdethUI.createFileInput('sub-file', 'Subtitle File (.srt, .ass, .vtt)', SUB_EXTS);
    form.appendChild(subFile);

    const burnToggle = EdethUI.createToggle('sub-burn', 'Hardsub: Burn subtitles permanently into video');
    form.appendChild(burnToggle);
    form.appendChild(EdethUI.createInfoText('Hardsub permanently draws text onto pixels. Softsub (unchecked) embeds the subtitle as a selectable stream inside MP4/MKV in under a second.'));

    const outputFile = EdethUI.createFileInput('sub-output', 'Output File (auto-generated if empty)', VIDEO_EXTS);
    form.appendChild(outputFile);

    videoFile.getInput().addEventListener('change', () => {
      const val = videoFile.getValue();
      if (val && !outputFile.getValue()) {
        outputFile.setValue(EdethUI.generateOutputName(val, '_subtitled'));
      }
    });

    const progress = EdethUI.createProgressBar('sub-progress');
    const terminal = EdethUI.createTaskTerminal('sub-terminal');

    const submitBtn = EdethUI.createButton('Add Subtitles', async () => {
      const video = videoFile.getValue();
      const sub = subFile.getValue();
      const output = outputFile.getValue() || EdethUI.generateOutputName(video, '_subtitled');

      if (!video) { EdethUI.showToast('Please select a video file', 'error'); return; }
      if (!sub) { EdethUI.showToast('Please select a subtitle file', 'error'); return; }

      submitBtn.disabled = true;
      outputFile.setValue(output);

      const body = document.getElementById('sub-terminal-body');
      if (body) body.innerHTML = '';

      try {
        const data = await this.apiCall('/api/subtitle/add', {
          videoFile: video,
          subtitleFile: sub,
          outputFile: output,
          burnIn: burnToggle.getValue()
        });
        this.subscribeProgress(data.taskId, 'sub-progress', 'sub-terminal', (res) => {
          submitBtn.disabled = false;
          if (res && res.status === 'complete') {
            const oldP = page.querySelector('#sub-preview');
            if (oldP) oldP.remove();
            const player = EdethUI.createVideoPlayer('sub-preview', output, 'Subtitled Video Preview');
            page.appendChild(player);
          }
        });
      } catch (err) {
        submitBtn.disabled = false;
      }
    }, 'primary', 'lg');

    form.appendChild(document.createElement('br'));
    form.appendChild(submitBtn);
    form.appendChild(progress);
    form.appendChild(terminal);

    const card = EdethUI.createCard('Subtitle Settings', form);
    page.appendChild(card);
    return page;
  },

  // ==============================
  // TRIM PAGE
  // ==============================
  renderTrim() {
    const page = document.createElement('div');

    page.appendChild(EdethUI.createPageHeader(
      'Trim Audio / Video',
      'Cut media files to exact timestamps using fast stream copy.'
    ));

    const form = document.createElement('div');

    const inputFile = EdethUI.createFileInput('trim-input', 'Input File (Video or Audio)', MEDIA_EXTS);
    form.appendChild(inputFile);

    const timeRow = document.createElement('div');
    timeRow.className = 'time-inputs-row';

    const startTime = EdethUI.createTimeInput('trim-start', 'Start Time (e.g. 00:00:10 or 10)', '00:00:00');
    const endTime = EdethUI.createTimeInput('trim-end', 'End Time (e.g. 00:00:30 or 30)', '00:00:00');
    timeRow.appendChild(startTime);
    timeRow.appendChild(endTime);
    form.appendChild(timeRow);

    const outputFile = EdethUI.createFileInput('trim-output', 'Output File (auto-generated if empty)', MEDIA_EXTS);
    form.appendChild(outputFile);

    inputFile.getInput().addEventListener('change', () => {
      const val = inputFile.getValue();
      if (val && !outputFile.getValue()) {
        outputFile.setValue(EdethUI.generateOutputName(val, '_trimmed'));
      }
    });

    const progress = EdethUI.createProgressBar('trim-progress');
    const terminal = EdethUI.createTaskTerminal('trim-terminal');

    const submitBtn = EdethUI.createButton('Trim Media', async () => {
      const input = inputFile.getValue();
      const start = startTime.getValue();
      const end = endTime.getValue();
      const output = outputFile.getValue() || EdethUI.generateOutputName(input, '_trimmed');

      if (!input) { EdethUI.showToast('Please select an input file', 'error'); return; }
      if (!start || !end) { EdethUI.showToast('Please enter both start and end times', 'error'); return; }

      submitBtn.disabled = true;
      outputFile.setValue(output);

      const body = document.getElementById('trim-terminal-body');
      if (body) body.innerHTML = '';

      try {
        const data = await this.apiCall('/api/trim', {
          inputFile: input,
          outputFile: output,
          startTime: start,
          endTime: end
        });
        this.subscribeProgress(data.taskId, 'trim-progress', 'trim-terminal', (res) => {
          submitBtn.disabled = false;
          if (res && res.status === 'complete') {
            const oldP = page.querySelector('#trim-preview');
            if (oldP) oldP.remove();
            const isAudio = AUDIO_EXTS.some(ext => output.toLowerCase().endsWith(ext));
            const player = isAudio ? EdethUI.createAudioPlayer('trim-preview', output, 'Trimmed Audio Preview')
                                   : EdethUI.createVideoPlayer('trim-preview', output, 'Trimmed Video Preview');
            page.appendChild(player);
          }
        });
      } catch (err) {
        submitBtn.disabled = false;
      }
    }, 'primary', 'lg');

    form.appendChild(document.createElement('br'));
    form.appendChild(submitBtn);
    form.appendChild(progress);
    form.appendChild(terminal);

    const card = EdethUI.createCard('Trim Settings', form);
    page.appendChild(card);
    return page;
  },

  // ==============================
  // EXTRACT AUDIO PAGE
  // ==============================
  renderExtract() {
    const page = document.createElement('div');

    page.appendChild(EdethUI.createPageHeader(
      'Extract Audio',
      'Separate the audio track from a video file without re-encoding video frames.'
    ));

    const form = document.createElement('div');

    const videoFile = EdethUI.createFileInput('ext-video', 'Video File', VIDEO_EXTS);
    form.appendChild(videoFile);

    const formatSelect = EdethUI.createSelect('ext-format', 'Output Format', [
      { value: 'mp3', text: 'MP3 (.mp3)' },
      { value: 'wav', text: 'WAV (.wav)' },
      { value: 'aac', text: 'AAC (.aac)' },
      { value: 'flac', text: 'FLAC (.flac)' },
      { value: 'ogg', text: 'OGG (.ogg)' }
    ]);
    form.appendChild(formatSelect);

    const outputFile = EdethUI.createFileInput('ext-output', 'Output File (auto-generated if empty)', AUDIO_EXTS);
    form.appendChild(outputFile);

    const updateOutput = () => {
      const val = videoFile.getValue();
      const fmt = formatSelect.getValue();
      if (val) {
        outputFile.setValue(EdethUI.generateOutputName(val, '_audio', '.' + fmt));
      }
    };
    videoFile.getInput().addEventListener('change', updateOutput);
    formatSelect.querySelector('select').addEventListener('change', updateOutput);

    const progress = EdethUI.createProgressBar('ext-progress');
    const terminal = EdethUI.createTaskTerminal('ext-terminal');

    const submitBtn = EdethUI.createButton('Extract Audio', async () => {
      const video = videoFile.getValue();
      const format = formatSelect.getValue();
      const output = outputFile.getValue() || EdethUI.generateOutputName(video, '_audio', '.' + format);

      if (!video) { EdethUI.showToast('Please select a video file', 'error'); return; }

      submitBtn.disabled = true;
      outputFile.setValue(output);

      const body = document.getElementById('ext-terminal-body');
      if (body) body.innerHTML = '';

      try {
        const data = await this.apiCall('/api/audio/extract', {
          videoFile: video,
          outputFile: output,
          format: format
        });
        this.subscribeProgress(data.taskId, 'ext-progress', 'ext-terminal', (res) => {
          submitBtn.disabled = false;
          if (res && res.status === 'complete') {
            const oldP = page.querySelector('#ext-preview');
            if (oldP) oldP.remove();
            const player = EdethUI.createAudioPlayer('ext-preview', output, 'Extracted Audio Preview');
            page.appendChild(player);
          }
        });
      } catch (err) {
        submitBtn.disabled = false;
      }
    }, 'primary', 'lg');

    form.appendChild(document.createElement('br'));
    form.appendChild(submitBtn);
    form.appendChild(progress);
    form.appendChild(terminal);

    const card = EdethUI.createCard('Extract Settings', form);
    page.appendChild(card);
    return page;
  },

  // ==============================
  // MERGE AUDIO PAGE
  // ==============================
  renderMerge() {
    const page = document.createElement('div');

    page.appendChild(EdethUI.createPageHeader(
      'Merge Audio into Video',
      'Add an audio track to a video file. Loop the audio to match video duration or insert at a specific timestamp.'
    ));

    const form = document.createElement('div');

    const videoFile = EdethUI.createFileInput('merge-video', 'Video File', VIDEO_EXTS);
    form.appendChild(videoFile);

    const audioFile = EdethUI.createFileInput('merge-audio', 'Audio File', AUDIO_EXTS);
    form.appendChild(audioFile);

    const loopToggle = EdethUI.createToggle('merge-loop', 'Loop audio across entire video duration');
    form.appendChild(loopToggle);

    const startTime = EdethUI.createTimeInput('merge-start', 'Or start audio at timestamp', '00:00:00');
    form.appendChild(startTime);
    form.appendChild(EdethUI.createInfoText('Leave as 00:00:00 to start at beginning. Automatically disabled when loop is enabled.'));

    const startInput = startTime.querySelector('input');
    loopToggle.getCheckbox().addEventListener('change', () => {
      if (loopToggle.getValue()) {
        startInput.disabled = true;
        startInput.value = '00:00:00';
        startInput.style.opacity = '0.5';
      } else {
        startInput.disabled = false;
        startInput.style.opacity = '1';
      }
    });

    const outputFile = EdethUI.createFileInput('merge-output', 'Output File (auto-generated if empty)', VIDEO_EXTS);
    form.appendChild(outputFile);

    videoFile.getInput().addEventListener('change', () => {
      const val = videoFile.getValue();
      if (val && !outputFile.getValue()) {
        outputFile.setValue(EdethUI.generateOutputName(val, '_merged'));
      }
    });

    const progress = EdethUI.createProgressBar('merge-progress');
    const terminal = EdethUI.createTaskTerminal('merge-terminal');

    const submitBtn = EdethUI.createButton('Merge Audio', async () => {
      const video = videoFile.getValue();
      const audio = audioFile.getValue();
      const output = outputFile.getValue() || EdethUI.generateOutputName(video, '_merged');

      if (!video) { EdethUI.showToast('Please select a video file', 'error'); return; }
      if (!audio) { EdethUI.showToast('Please select an audio file', 'error'); return; }

      submitBtn.disabled = true;
      outputFile.setValue(output);

      let startAt = 0;
      if (!loopToggle.getValue()) {
        const timeStr = startTime.getValue() || '00:00:00';
        const parts = timeStr.split(':');
        if (parts.length === 3) {
          startAt = parseInt(parts[0]) * 3600 + parseInt(parts[1]) * 60 + parseFloat(parts[2]);
        } else if (parts.length === 2) {
          startAt = parseInt(parts[0]) * 60 + parseFloat(parts[1]);
        } else if (parts.length === 1) {
          startAt = parseFloat(parts[0]) || 0;
        }
      }

      const body = document.getElementById('merge-terminal-body');
      if (body) body.innerHTML = '';

      try {
        const data = await this.apiCall('/api/audio/merge', {
          videoFile: video,
          audioFile: audio,
          outputFile: output,
          loop: loopToggle.getValue(),
          startAt: startAt
        });
        this.subscribeProgress(data.taskId, 'merge-progress', 'merge-terminal', (res) => {
          submitBtn.disabled = false;
          if (res && res.status === 'complete') {
            const oldP = page.querySelector('#merge-preview');
            if (oldP) oldP.remove();
            const player = EdethUI.createVideoPlayer('merge-preview', output, 'Merged Video Preview');
            page.appendChild(player);
          }
        });
      } catch (err) {
        submitBtn.disabled = false;
      }
    }, 'primary', 'lg');

    form.appendChild(document.createElement('br'));
    form.appendChild(submitBtn);
    form.appendChild(progress);
    form.appendChild(terminal);

    const card = EdethUI.createCard('Merge Settings', form);
    page.appendChild(card);
    return page;
  },

  // ==============================
  // VIDEO COMPRESSOR PAGE
  // ==============================
  renderCompress() {
    const page = document.createElement('div');

    page.appendChild(EdethUI.createPageHeader(
      'Smart Video Compressor & Optimizer',
      'Target file size encoding for Discord (25MB), Telegram (50MB), or web streaming with real-time FFmpeg progress.'
    ));

    const form = document.createElement('div');

    const videoFile = EdethUI.createFileInput('comp-video', 'Source Video File', VIDEO_EXTS);
    form.appendChild(videoFile);

    // Preset Grid Selection
    const presetLabel = document.createElement('label');
    presetLabel.className = 'form-label';
    presetLabel.textContent = 'Compression Preset / Target Size';
    form.appendChild(presetLabel);

    const presetGrid = document.createElement('div');
    presetGrid.className = 'preset-grid';

    const presets = [
      { id: 'discord25', title: 'Discord', badge: '25 MB', desc: 'Fits Discord file size limit without Nitro' },
      { id: 'telegram50', title: 'Telegram', badge: '50 MB', desc: 'Fits Telegram fast streaming upload limit' },
      { id: 'web720p', title: 'Web 720p', badge: '720p HD', desc: 'Crisp HD streaming video, reduced bitrate' },
      { id: 'web1080p', title: 'Web 1080p', badge: '1080p FHD', desc: 'High quality web standard' },
      { id: 'ultracompact', title: 'Ultra Compact', badge: 'CRF 28', desc: 'Maximum compression for archiving' },
      { id: 'custom', title: 'Custom Size', badge: 'Target MB', desc: 'Specify exact target megabytes' }
    ];

    let selectedPreset = 'discord25';

    presets.forEach(p => {
      const card = document.createElement('div');
      card.className = 'preset-card' + (p.id === selectedPreset ? ' active' : '');
      card.dataset.preset = p.id;
      card.innerHTML = `
        <div class="preset-card-title">
          <span>${p.title}</span>
          <span class="preset-card-badge">${p.badge}</span>
        </div>
        <div class="preset-card-desc">${p.desc}</div>
      `;
      card.onclick = () => {
        presetGrid.querySelectorAll('.preset-card').forEach(c => c.classList.remove('active'));
        card.classList.add('active');
        selectedPreset = p.id;
        customSizeGroup.style.display = p.id === 'custom' ? 'block' : 'none';
      };
      presetGrid.appendChild(card);
    });
    form.appendChild(presetGrid);

    // Custom Target Size Input (initially hidden)
    const customSizeGroup = document.createElement('div');
    customSizeGroup.className = 'form-group';
    customSizeGroup.style.display = 'none';
    customSizeGroup.innerHTML = `
      <label class="form-label">Custom Target Size (MB)</label>
      <input type="number" id="comp-custom-size" class="form-control" placeholder="e.g. 15 or 30" min="1" max="2000" step="0.5" value="20">
    `;
    form.appendChild(customSizeGroup);

    // Advanced Options (Resolution & Codec)
    const optionsRow = document.createElement('div');
    optionsRow.style.display = 'grid';
    optionsRow.style.gridTemplateColumns = 'repeat(auto-fit, minmax(200px, 1fr))';
    optionsRow.style.gap = '14px';
    optionsRow.style.marginTop = '10px';

    const resSelect = EdethUI.createSelect('comp-res', 'Resolution', [
      { value: 'original', text: 'Keep Original Resolution' },
      { value: '1080p', text: '1080p Full HD (1920x1080)' },
      { value: '720p', text: '720p HD (1280x720)' },
      { value: '480p', text: '480p SD (854x480)' },
      { value: '360p', text: '360p Mobile (640x360)' }
    ]);

    const codecSelect = EdethUI.createSelect('comp-codec', 'Video Codec', [
      { value: 'h264', text: 'H.264 / AVC (Universally Compatible)' },
      { value: 'h265', text: 'H.265 / HEVC (30% Smaller, Modern Devices)' }
    ]);

    optionsRow.appendChild(resSelect);
    optionsRow.appendChild(codecSelect);
    form.appendChild(optionsRow);

    const outputFile = EdethUI.createFileInput('comp-output', 'Output Compressed Video', VIDEO_EXTS);
    form.appendChild(outputFile);

    videoFile.getInput().addEventListener('change', () => {
      const val = videoFile.getValue();
      if (val && !outputFile.getValue()) {
        outputFile.setValue(EdethUI.generateOutputName(val, '_compressed', '.mp4'));
      }
    });

    const progress = EdethUI.createProgressBar('comp-progress');
    const terminal = EdethUI.createTaskTerminal('comp-terminal');

    const submitBtn = EdethUI.createButton('Compress Video', async () => {
      const video = videoFile.getValue();
      const output = outputFile.getValue() || EdethUI.generateOutputName(video, '_compressed', '.mp4');

      if (!video) { EdethUI.showToast('Please select a video file to compress', 'error'); return; }

      let targetMB = 0;
      if (selectedPreset === 'discord25') targetMB = 24.0;
      else if (selectedPreset === 'telegram50') targetMB = 48.5;
      else if (selectedPreset === 'custom') {
        const customInput = customSizeGroup.querySelector('input');
        targetMB = parseFloat(customInput.value) || 20;
      }

      submitBtn.disabled = true;
      outputFile.setValue(output);

      const body = document.getElementById('comp-terminal-body');
      if (body) body.innerHTML = '';

      try {
        const data = await this.apiCall('/api/video/compress', {
          videoFile: video,
          outputFile: output,
          preset: selectedPreset,
          targetSizeMB: targetMB,
          resolution: resSelect.getValue(),
          codec: codecSelect.getValue()
        });

        this.subscribeProgress(data.taskId, 'comp-progress', 'comp-terminal', (res) => {
          submitBtn.disabled = false;
          if (res && res.status === 'complete') {
            const oldP = page.querySelector('#comp-preview');
            if (oldP) oldP.remove();
            const player = EdethUI.createVideoPlayer('comp-preview', output, 'Compressed Video Preview');
            page.appendChild(player);
          }
        });
      } catch (err) {
        submitBtn.disabled = false;
      }
    }, 'primary', 'lg');

    form.appendChild(document.createElement('br'));
    form.appendChild(submitBtn);
    form.appendChild(progress);
    form.appendChild(terminal);

    const card = EdethUI.createCard('Compression Settings', form);
    page.appendChild(card);
    return page;
  },

  // ==============================
  // SSE Real-Time Progress & Log Subscription
  // ==============================
  subscribeProgress(taskId, progressBarId, terminalId, onComplete) {
    const evtSource = new EventSource('/api/progress/' + taskId);

    evtSource.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        EdethUI.updateProgress(progressBarId, data.progress, data.message);

        if (data.log && terminalId) {
          EdethUI.appendTerminalLog(terminalId, data.log);
        }

        if (data.status === 'complete') {
          evtSource.close();
          const outMsg = data.outputFile ? 'Saved to: ' + data.outputFile : 'Operation completed successfully!';
          EdethUI.showToast(outMsg, 'success', 6000);
          if (terminalId) {
            EdethUI.appendTerminalLog(terminalId, '==> [DONE] ' + outMsg);
          }
          if (onComplete) onComplete(data);
        } else if (data.status === 'error') {
          evtSource.close();
          EdethUI.showToast(data.message, 'error', 6000);
          if (terminalId) {
            EdethUI.appendTerminalLog(terminalId, '==> [FAILED] ' + data.message);
          }
          if (onComplete) onComplete(data);
        }
      } catch (err) {
        console.error('Failed to parse SSE data:', err);
      }
    };

    evtSource.onerror = () => {
      evtSource.close();
      const container = document.getElementById(progressBarId);
      if (container && !container.classList.contains('complete')) {
        EdethUI.showToast('Background process finished or connection closed.', 'info', 3000);
      }
      if (onComplete) onComplete(null);
    };
  },

  // ==============================
  // API Call Helper
  // ==============================
  async apiCall(endpoint, body) {
    try {
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || 'Request failed');
      }

      return data;
    } catch (err) {
      EdethUI.showToast(err.message, 'error', 5000);
      throw err;
    }
  }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => App.init());
