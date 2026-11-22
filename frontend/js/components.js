/* ============================================
   Edeth UI Components
   ============================================ */

const EdethUI = {

  // ---- SVG Icon helpers ----
  icons: {
    home: '<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>',
    wand: '<path d="M15 4V2"/><path d="M15 16v-2"/><path d="M8 9h2"/><path d="M20 9h2"/><path d="M17.8 11.8L19 13"/><path d="M15 9h0"/><path d="M17.8 6.2L19 5"/><path d="M3 21l9-9"/><path d="M12.2 6.2L11 5"/>',
    subtitle: '<rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="14" x2="23" y2="14"/><line x1="5" y1="18" x2="12" y2="18"/><line x1="15" y1="18" x2="19" y2="18"/>',
    scissors: '<circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><line x1="20" y1="4" x2="8.12" y2="15.88"/><line x1="14.47" y1="14.48" x2="20" y2="20"/><line x1="8.12" y1="8.12" x2="12" y2="12"/>',
    music: '<path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>',
    merge: '<path d="M8 6L21 6"/><path d="M8 12L21 12"/><path d="M8 18L21 18"/><path d="M3 6L3 18"/><path d="M3 6L8 12"/><path d="M3 18L8 12"/>',
    folder: '<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>',
    file: '<path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/>',
    musicNote: '<path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>',
    film: '<rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="2" y1="7" x2="7" y2="7"/><line x1="2" y1="17" x2="7" y2="17"/><line x1="17" y1="7" x2="22" y2="7"/><line x1="17" y1="17" x2="22" y2="17"/>',
    text: '<polyline points="4 7 4 4 20 4 20 7"/><line x1="9" y1="20" x2="15" y2="20"/><line x1="12" y1="4" x2="12" y2="20"/>',
    drive: '<line x1="22" y1="12" x2="2" y2="12"/><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/><line x1="6" y1="16" x2="6.01" y2="16"/><line x1="10" y1="16" x2="10.01" y2="16"/>',
    check: '<polyline points="20 6 9 17 4 12"/>',
    x: '<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>',
    alertCircle: '<circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>',
    info: '<circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>',
    upload: '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>',
    compress: '<path d="M4 14h6m-6 0l3-3m-3 3l3 3"/><path d="M20 10h-6m6 0l-3-3m3 3l-3 3"/><rect x="2" y="3" width="20" height="18" rx="2"/>',
    folderOpen: '<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>'
  },

  createSVG(iconName, size = 20) {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('width', size);
    svg.setAttribute('height', size);
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '2');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('stroke-linejoin', 'round');
    svg.innerHTML = this.icons[iconName] || this.icons.file;
    return svg;
  },

  cleanPathString(str) {
    if (!str) return '';
    return str.trim().replace(/^["'`]|["'`]$/g, '').trim();
  },

  // ---- Card component ----
  createCard(title, content) {
    const card = document.createElement('div');
    card.className = 'card';
    if (title) {
      const header = document.createElement('div');
      header.className = 'card-header';
      const titleEl = document.createElement('h3');
      titleEl.className = 'card-title';
      titleEl.textContent = title;
      header.appendChild(titleEl);
      card.appendChild(header);
    }
    if (content) {
      if (typeof content === 'string') {
        const desc = document.createElement('p');
        desc.className = 'card-description';
        desc.textContent = content;
        card.appendChild(desc);
      } else {
        card.appendChild(content);
      }
    }
    return card;
  },

  // ---- Feature card for home page ----
  createFeatureCard(title, description, iconName, page, isMain = false) {
    const card = document.createElement('div');
    card.className = 'feature-card' + (isMain ? ' main-feature' : '');
    card.onclick = () => { window.location.hash = '#' + page; };

    const iconWrapper = document.createElement('div');
    iconWrapper.className = 'feature-icon';
    iconWrapper.appendChild(this.createSVG(iconName, isMain ? 32 : 24));
    card.appendChild(iconWrapper);

    const textContainer = document.createElement('div');

    if (isMain) {
      const badge = document.createElement('span');
      badge.className = 'feature-badge';
      badge.textContent = '⭐ MAIN FEATURE';
      textContainer.appendChild(badge);
    }

    const titleEl = document.createElement('div');
    titleEl.className = 'feature-title';
    titleEl.textContent = title;
    textContainer.appendChild(titleEl);

    const descEl = document.createElement('div');
    descEl.className = 'feature-desc';
    descEl.textContent = description;
    textContainer.appendChild(descEl);

    card.appendChild(textContainer);
    return card;
  },

  // ---- File input with browse button ----
  createFileInput(id, label, acceptExts = []) {
    const group = document.createElement('div');
    group.className = 'form-group';

    const labelEl = document.createElement('label');
    labelEl.setAttribute('for', id);
    labelEl.textContent = label;
    group.appendChild(labelEl);

    const inputGroup = document.createElement('div');
    inputGroup.className = 'file-input-group';

    const input = document.createElement('input');
    input.type = 'text';
    input.id = id;
    input.name = id;
    input.placeholder = 'Select a file or paste path...';

    // Auto-strip surrounding quotes when typing or pasting
    input.addEventListener('input', () => {
      const cleaned = EdethUI.cleanPathString(input.value);
      if (cleaned !== input.value) {
        input.value = cleaned;
      }
    });

    inputGroup.appendChild(input);

    const browseBtn = document.createElement('button');
    browseBtn.type = 'button';
    browseBtn.className = 'btn btn-secondary';
    browseBtn.innerHTML = `
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:2px;">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
      </svg>
      <span>Browse</span>
    `;
    browseBtn.onclick = async () => {
      const result = await FileBrowser.open({
        accept: acceptExts,
        title: label
      });
      if (result) {
        input.value = EdethUI.cleanPathString(result);
        input.dispatchEvent(new Event('change', { bubbles: true }));
      }
    };
    inputGroup.appendChild(browseBtn);

    // Upload button for Docker / remote client streaming upload
    const uploadBtn = document.createElement('button');
    uploadBtn.type = 'button';
    uploadBtn.className = 'btn btn-secondary';
    uploadBtn.innerHTML = `
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:2px;">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
        <polyline points="17 8 12 3 7 8"></polyline>
        <line x1="12" y1="3" x2="12" y2="15"></line>
      </svg>
      <span>Upload</span>
    `;
    uploadBtn.title = 'Upload file from device to server (Docker/Cloud mode)';

    const hiddenFileInput = document.createElement('input');
    hiddenFileInput.type = 'file';
    hiddenFileInput.style.display = 'none';
    if (acceptExts && acceptExts.length > 0) {
      hiddenFileInput.accept = acceptExts.join(',');
    }

    uploadBtn.onclick = () => hiddenFileInput.click();

    hiddenFileInput.onchange = async () => {
      const file = hiddenFileInput.files[0];
      if (!file) return;

      uploadBtn.disabled = true;
      uploadBtn.textContent = 'Uploading...';

      try {
        const formData = new FormData();
        formData.append('file', file);

        const res = await fetch('/api/upload', {
          method: 'POST',
          body: formData
        });

        const data = await res.json();
        if (!res.ok || data.error) {
          throw new Error(data.error || 'Upload failed');
        }

        input.value = EdethUI.cleanPathString(data.filePath);
        input.dispatchEvent(new Event('change', { bubbles: true }));
        EdethUI.showToast(`Uploaded "${file.name}" to server temp storage`, 'success', 3500);
      } catch (err) {
        EdethUI.showToast(`Upload failed: ${err.message}`, 'error', 5000);
      } finally {
        uploadBtn.disabled = false;
        uploadBtn.textContent = 'Upload';
        hiddenFileInput.value = '';
      }
    };

    inputGroup.appendChild(uploadBtn);
    inputGroup.appendChild(hiddenFileInput);

    group.appendChild(inputGroup);

    group.getValue = () => EdethUI.cleanPathString(input.value);
    group.setValue = (val) => { input.value = EdethUI.cleanPathString(val); };
    group.getInput = () => input;
    return group;
  },

  // ---- Button ----
  createButton(text, onClick, type = 'primary', size = '') {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'btn btn-' + type + (size ? ' btn-' + size : '');
    btn.textContent = text;
    if (onClick) btn.onclick = onClick;
    return btn;
  },

  // ---- Progress bar ----
  createProgressBar(id) {
    const container = document.createElement('div');
    container.className = 'progress-container hidden';
    container.id = id;

    const textRow = document.createElement('div');
    textRow.className = 'progress-text';

    const label = document.createElement('span');
    label.className = 'progress-label';
    label.textContent = 'Processing...';
    textRow.appendChild(label);

    const percent = document.createElement('span');
    percent.className = 'progress-percent';
    percent.textContent = '0%';
    textRow.appendChild(percent);

    container.appendChild(textRow);

    const wrapper = document.createElement('div');
    wrapper.className = 'progress-bar-wrapper';

    const fill = document.createElement('div');
    fill.className = 'progress-bar-fill';
    fill.style.width = '0%';
    wrapper.appendChild(fill);

    container.appendChild(wrapper);
    return container;
  },

  updateProgress(id, percentVal, message = '') {
    const container = document.getElementById(id);
    if (!container) return;

    container.classList.remove('hidden');
    const fill = container.querySelector('.progress-bar-fill');
    const percentEl = container.querySelector('.progress-percent');
    const labelEl = container.querySelector('.progress-label');

    if (fill) fill.style.width = Math.min(percentVal, 100).toFixed(1) + '%';
    if (percentEl) percentEl.textContent = Math.round(percentVal) + '%';
    if (labelEl && message) labelEl.textContent = message;

    if (percentVal >= 100) {
      container.classList.add('complete');
    } else {
      container.classList.remove('complete');
    }
  },

  // ---- Task real-time log terminal ----
  createTaskTerminal(id) {
    const term = document.createElement('div');
    term.className = 'task-log-terminal hidden';
    term.id = id;

    const header = document.createElement('div');
    header.className = 'terminal-header';

    const leftBox = document.createElement('div');
    leftBox.style.display = 'flex';
    leftBox.style.alignItems = 'center';
    leftBox.style.gap = '10px';

    const dots = document.createElement('div');
    dots.style.display = 'flex';
    dots.style.gap = '5px';
    dots.innerHTML = `
      <span style="width:8px; height:8px; border-radius:50%; background:#ff5f56; display:inline-block; opacity:0.8;"></span>
      <span style="width:8px; height:8px; border-radius:50%; background:#ffbd2e; display:inline-block; opacity:0.8;"></span>
      <span style="width:8px; height:8px; border-radius:50%; background:#27c93f; display:inline-block; opacity:0.8;"></span>
    `;
    leftBox.appendChild(dots);

    const title = document.createElement('span');
    title.className = 'terminal-title';
    title.textContent = 'Live Execution Log';
    leftBox.appendChild(title);
    header.appendChild(leftBox);

    const toggleBtn = document.createElement('button');
    toggleBtn.type = 'button';
    toggleBtn.className = 'terminal-toggle-btn';
    toggleBtn.textContent = 'Collapse';
    toggleBtn.onclick = () => {
      term.classList.toggle('collapsed');
      toggleBtn.textContent = term.classList.contains('collapsed') ? 'Expand' : 'Collapse';
    };
    header.appendChild(toggleBtn);

    term.appendChild(header);

    const body = document.createElement('div');
    body.className = 'terminal-body';
    body.id = id + '-body';
    term.appendChild(body);

    return term;
  },

  appendTerminalLog(id, line) {
    if (!line) return;
    const term = document.getElementById(id);
    if (!term) return;

    term.classList.remove('hidden');
    const body = document.getElementById(id + '-body') || term;
    const lineEl = document.createElement('div');
    lineEl.className = 'terminal-line';

    const lower = line.toLowerCase();
    if (lower.includes('error') || lower.includes('failed')) {
      lineEl.className += ' error';
    } else if (lower.includes('complete') || lower.includes('success')) {
      lineEl.className += ' info';
    }

    lineEl.textContent = line;
    body.appendChild(lineEl);

    // Limit buffer length
    if (body.children.length > 250) {
      body.removeChild(body.firstChild);
    }
    body.scrollTop = body.scrollHeight;
  },

  // ---- Select dropdown ----
  createSelect(id, label, options) {
    const group = document.createElement('div');
    group.className = 'form-group';

    const labelEl = document.createElement('label');
    labelEl.setAttribute('for', id);
    labelEl.textContent = label;
    group.appendChild(labelEl);

    const select = document.createElement('select');
    select.id = id;
    select.name = id;

    options.forEach(opt => {
      const option = document.createElement('option');
      option.value = typeof opt === 'object' ? opt.value : opt;
      option.textContent = typeof opt === 'object' ? opt.text : opt;
      select.appendChild(option);
    });

    group.appendChild(select);
    group.getValue = () => select.value;
    return group;
  },

  // ---- Time input ----
  createTimeInput(id, label, placeholder = '00:00:00') {
    const group = document.createElement('div');
    group.className = 'form-group';

    const labelEl = document.createElement('label');
    labelEl.setAttribute('for', id);
    labelEl.textContent = label;
    group.appendChild(labelEl);

    const input = document.createElement('input');
    input.type = 'text';
    input.id = id;
    input.name = id;
    input.className = 'time-input';
    input.placeholder = placeholder;
    group.appendChild(input);

    group.getValue = () => input.value.trim();
    group.setValue = (v) => { input.value = v; };
    return group;
  },

  // ---- Toggle switch ----
  createToggle(id, label, defaultChecked = false) {
    const group = document.createElement('div');
    group.className = 'form-group';

    const container = document.createElement('div');
    container.className = 'toggle-container';

    const toggle = document.createElement('label');
    toggle.className = 'toggle';

    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.id = id;
    checkbox.name = id;
    checkbox.checked = defaultChecked;
    toggle.appendChild(checkbox);

    const slider = document.createElement('span');
    slider.className = 'toggle-slider';
    toggle.appendChild(slider);

    container.appendChild(toggle);

    const labelText = document.createElement('span');
    labelText.className = 'toggle-label-text';
    labelText.textContent = label;
    labelText.onclick = () => { checkbox.checked = !checkbox.checked; checkbox.dispatchEvent(new Event('change')); };
    container.appendChild(labelText);

    group.appendChild(container);
    group.getValue = () => checkbox.checked;
    group.getCheckbox = () => checkbox;
    return group;
  },

  // ---- Info text ----
  createInfoText(text) {
    const info = document.createElement('p');
    info.className = 'info-text';
    info.textContent = text;
    return info;
  },

  // ---- Toast notifications ----
  showToast(message, type = 'success', duration = 4000) {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = 'toast toast-' + type;

    const iconDiv = document.createElement('div');
    iconDiv.className = 'toast-icon';
    if (type === 'success') {
      iconDiv.appendChild(this.createSVG('check', 20));
    } else if (type === 'error') {
      iconDiv.appendChild(this.createSVG('alertCircle', 20));
    } else {
      iconDiv.appendChild(this.createSVG('info', 20));
    }
    toast.appendChild(iconDiv);

    const msgDiv = document.createElement('div');
    msgDiv.className = 'toast-message';
    msgDiv.textContent = message;
    toast.appendChild(msgDiv);

    container.appendChild(toast);

    setTimeout(() => {
      toast.classList.add('slide-out');
      setTimeout(() => toast.remove(), 300);
    }, duration);
  },

  // ---- Utility formatters ----
  formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  },

  formatDuration(seconds) {
    if (!seconds || seconds <= 0) return '00:00:00';
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    return [h, m, s].map(v => String(v).padStart(2, '0')).join(':');
  },

  // ---- Generate output filename ----
  generateOutputName(inputPath, suffix = '_output', newExt = null) {
    if (!inputPath) return '';
    inputPath = this.cleanPathString(inputPath);
    const lastDot = inputPath.lastIndexOf('.');
    const lastSlash = Math.max(inputPath.lastIndexOf('/'), inputPath.lastIndexOf('\\'));
    const dir = inputPath.substring(0, lastSlash + 1);
    const name = lastDot > lastSlash ? inputPath.substring(lastSlash + 1, lastDot) : inputPath.substring(lastSlash + 1);
    const ext = newExt || (lastDot > lastSlash ? inputPath.substring(lastDot) : '');
    return dir + name + suffix + ext;
  },

  // ---- Create page header ----
  createPageHeader(title, subtitle) {
    const header = document.createElement('div');
    header.className = 'page-header';

    const titleEl = document.createElement('h1');
    titleEl.className = 'page-title';
    titleEl.textContent = title;
    header.appendChild(titleEl);

    if (subtitle) {
      const subEl = document.createElement('p');
      subEl.className = 'page-subtitle';
      subEl.textContent = subtitle;
      header.appendChild(subEl);
    }

    return header;
  },

  // ---- Audio Preview Player ----
  createAudioPlayer(id, filePath, title = 'Audio Preview') {
    const card = document.createElement('div');
    card.className = 'media-player-card';
    card.id = id;

    const streamUrl = '/api/stream?path=' + encodeURIComponent(EdethUI.cleanPathString(filePath));

    card.innerHTML = `
      <div class="media-player-header">
        <div class="media-player-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"></polygon>
            <path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"></path>
          </svg>
          <span>${title}</span>
        </div>
        <span class="file-name" style="font-size:0.8rem; color:var(--text-muted);">${EdethUI.cleanPathString(filePath).split(/[/\\]/).pop()}</span>
      </div>
      <div class="custom-audio-player">
        <audio src="${streamUrl}" preload="metadata"></audio>
        <button type="button" class="player-btn-circle" title="Play">▶</button>
        <div class="audio-progress-bar">
          <div class="audio-progress-fill"></div>
        </div>
        <span class="player-time-display">00:00 / 00:00</span>
      </div>
    `;

    const audio = card.querySelector('audio');
    const playBtn = card.querySelector('.player-btn-circle');
    const progressBar = card.querySelector('.audio-progress-bar');
    const progressFill = card.querySelector('.audio-progress-fill');
    const timeDisplay = card.querySelector('.player-time-display');

    playBtn.onclick = () => {
      if (audio.paused) {
        audio.play().catch(e => console.warn(e));
        playBtn.textContent = '⏸';
        playBtn.title = 'Pause';
      } else {
        audio.pause();
        playBtn.textContent = '▶';
        playBtn.title = 'Play';
      }
    };

    audio.onended = () => {
      playBtn.textContent = '▶';
      progressFill.style.width = '0%';
    };

    audio.ontimeupdate = () => {
      if (!audio.duration) return;
      const pct = (audio.currentTime / audio.duration) * 100;
      progressFill.style.width = pct + '%';
      timeDisplay.textContent = `${EdethUI.formatDuration(audio.currentTime)} / ${EdethUI.formatDuration(audio.duration)}`;
    };

    audio.onloadedmetadata = () => {
      timeDisplay.textContent = `00:00 / ${EdethUI.formatDuration(audio.duration)}`;
    };

    progressBar.onclick = (e) => {
      if (!audio.duration) return;
      const rect = progressBar.getBoundingClientRect();
      const clickPct = (e.clientX - rect.left) / rect.width;
      audio.currentTime = clickPct * audio.duration;
    };

    return card;
  },

  // ---- Video Preview Player ----
  createVideoPlayer(id, filePath, title = 'Video Preview') {
    const card = document.createElement('div');
    card.className = 'media-player-card';
    card.id = id;

    const streamUrl = '/api/stream?path=' + encodeURIComponent(EdethUI.cleanPathString(filePath));

    card.innerHTML = `
      <div class="media-player-header">
        <div class="media-player-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="23 7 16 12 23 17 23 7"></polygon>
            <rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
          </svg>
          <span>${title}</span>
        </div>
        <span style="font-size:0.8rem; color:var(--text-muted);">${EdethUI.cleanPathString(filePath).split(/[/\\]/).pop()}</span>
      </div>
      <div class="video-preview-wrapper">
        <video controls preload="metadata" src="${streamUrl}"></video>
      </div>
    `;

    return card;
  },

  // ---- A/B Comparison Audio Player with Waveform & Watermark Markers ----
  createABComparisonPlayer(id, origPath, cleanPath, watermarkMarkers = []) {
    const card = document.createElement('div');
    card.className = 'media-player-card';
    card.id = id;

    const origStream = '/api/stream?path=' + encodeURIComponent(EdethUI.cleanPathString(origPath));
    const cleanStream = '/api/stream?path=' + encodeURIComponent(EdethUI.cleanPathString(cleanPath));

    card.innerHTML = `
      <div class="media-player-header">
        <div class="media-player-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 18V5l12-2v13"></path>
            <circle cx="6" cy="18" r="3"></circle>
            <circle cx="18" cy="16" r="3"></circle>
          </svg>
          <span>Interactive Audio Quality Comparison (A/B)</span>
        </div>
        <span style="font-size:0.8rem; color:var(--green-primary); font-weight:600;">Seamless Real-Time Switch</span>
      </div>

      <div class="ab-switcher-bar">
        <button type="button" class="ab-toggle-btn active" data-source="clean">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
          <span>Clean Output (Watermark Removed)</span>
        </button>
        <button type="button" class="ab-toggle-btn" data-source="orig">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line></svg>
          <span>Original (Watermarked)</span>
        </button>
      </div>

      <div class="waveform-canvas-container" title="Click to seek anywhere">
        <canvas height="90"></canvas>
        <div class="waveform-playhead"></div>
      </div>

      <div class="waveform-legend">
        <div class="waveform-legend-tag">
          <span class="waveform-legend-dot green"></span>
          <span>Waveform Playback</span>
        </div>
        <div class="waveform-legend-tag">
          <span class="waveform-legend-dot red"></span>
          <span>Detected Watermark Positions (${watermarkMarkers.length} removed)</span>
        </div>
        <div class="player-time-display ab-time">00:00 / 00:00</div>
      </div>

      <div class="custom-audio-player" style="margin-top:10px;">
        <audio id="${id}-audio-clean" src="${cleanStream}" preload="auto"></audio>
        <audio id="${id}-audio-orig" src="${origStream}" preload="auto" muted></audio>
        <button type="button" class="player-btn-circle" title="Play">▶</button>
        <div class="audio-progress-bar">
          <div class="audio-progress-fill"></div>
        </div>
        <div style="font-size:0.8rem; color:var(--text-muted); white-space:nowrap;">
          Tip: Toggle buttons while playing for instant A/B comparison!
        </div>
      </div>
    `;

    const audioClean = card.querySelector(`#${id}-audio-clean`);
    const audioOrig = card.querySelector(`#${id}-audio-orig`);
    const playBtn = card.querySelector('.player-btn-circle');
    const progressBar = card.querySelector('.audio-progress-bar');
    const progressFill = card.querySelector('.audio-progress-fill');
    const timeDisplay = card.querySelector('.ab-time');
    const canvasContainer = card.querySelector('.waveform-canvas-container');
    const canvas = card.querySelector('canvas');
    const playhead = card.querySelector('.waveform-playhead');
    const btnClean = card.querySelector('[data-source="clean"]');
    const btnOrig = card.querySelector('[data-source="orig"]');

    let currentSource = 'clean'; // 'clean' or 'orig'

    // Synchronize play/pause
    const togglePlay = () => {
      const activeAudio = currentSource === 'clean' ? audioClean : audioOrig;
      if (activeAudio.paused) {
        audioClean.currentTime = activeAudio.currentTime;
        audioOrig.currentTime = activeAudio.currentTime;
        audioClean.play().catch(e => console.warn(e));
        audioOrig.play().catch(e => console.warn(e));
        playBtn.textContent = '⏸';
      } else {
        audioClean.pause();
        audioOrig.pause();
        playBtn.textContent = '▶';
      }
    };
    playBtn.onclick = togglePlay;

    // Switch A / B
    btnClean.onclick = () => {
      if (currentSource === 'clean') return;
      currentSource = 'clean';
      btnClean.classList.add('active');
      btnOrig.classList.remove('active');
      audioClean.currentTime = audioOrig.currentTime;
      audioClean.muted = false;
      audioOrig.muted = true;
    };

    btnOrig.onclick = () => {
      if (currentSource === 'orig') return;
      currentSource = 'orig';
      btnOrig.classList.add('active');
      btnClean.classList.remove('active');
      audioOrig.currentTime = audioClean.currentTime;
      audioOrig.muted = false;
      audioClean.muted = true;
    };

    // Seek
    const seekTo = (seconds) => {
      audioClean.currentTime = seconds;
      audioOrig.currentTime = seconds;
    };

    progressBar.onclick = (e) => {
      const activeAudio = audioClean;
      if (!activeAudio.duration) return;
      const rect = progressBar.getBoundingClientRect();
      const pct = (e.clientX - rect.left) / rect.width;
      seekTo(pct * activeAudio.duration);
    };

    canvasContainer.onclick = (e) => {
      const activeAudio = audioClean;
      if (!activeAudio.duration) return;
      const rect = canvasContainer.getBoundingClientRect();
      const pct = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
      seekTo(pct * activeAudio.duration);
    };

    // Canvas drawing
    const drawWaveform = () => {
      const width = canvasContainer.clientWidth || 600;
      canvas.width = width;
      const height = canvas.height;
      const ctx = canvas.getContext('2d');
      ctx.clearRect(0, 0, width, height);

      // Draw subtle grid
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.04)';
      ctx.lineWidth = 1;
      for (let y = 16; y < height; y += 22) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
        ctx.stroke();
      }

      // Draw background center line
      ctx.strokeStyle = 'rgba(0, 230, 118, 0.2)';
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(0, height / 2);
      ctx.lineTo(width, height / 2);
      ctx.stroke();

      // Render aesthetic waveform bars
      const numBars = Math.floor(width / 4);
      const barWidth = 2.4;
      for (let i = 0; i < numBars; i++) {
        const x = i * 4;
        const seed = Math.sin(i * 0.14) * 0.45 + Math.cos(i * 0.06) * 0.35 + 0.45;
        const barH = Math.max(6, seed * (height - 22));
        const y = (height - barH) / 2;

        const grad = ctx.createLinearGradient(0, y, 0, y + barH);
        grad.addColorStop(0, '#00E676');
        grad.addColorStop(0.5, '#00C853');
        grad.addColorStop(1, '#007A38');
        ctx.fillStyle = grad;
        ctx.fillRect(x, y, barWidth, barH);
      }

      // Draw detected watermark markers
      const duration = audioClean.duration || 120;
      if (watermarkMarkers && watermarkMarkers.length > 0 && duration > 0) {
        watermarkMarkers.forEach(sec => {
          const x = (sec / duration) * width;
          // Red marker glow line
          ctx.strokeStyle = 'rgba(255, 82, 82, 0.3)';
          ctx.lineWidth = 4;
          ctx.beginPath();
          ctx.moveTo(x, 0);
          ctx.lineTo(x, height);
          ctx.stroke();

          // Sharp red line
          ctx.strokeStyle = '#FF5252';
          ctx.lineWidth = 1.5;
          ctx.beginPath();
          ctx.moveTo(x, 0);
          ctx.lineTo(x, height);
          ctx.stroke();

          // Small marker badge at top
          ctx.fillStyle = '#FF5252';
          ctx.beginPath();
          ctx.arc(x, 7, 4, 0, Math.PI * 2);
          ctx.fill();
        });
      }
    };

    audioClean.onloadedmetadata = () => {
      timeDisplay.textContent = `00:00 / ${EdethUI.formatDuration(audioClean.duration)}`;
      drawWaveform();
    };

    window.addEventListener('resize', drawWaveform);

    audioClean.ontimeupdate = () => {
      const dur = audioClean.duration;
      if (!dur) return;
      const cur = audioClean.currentTime;
      const pct = (cur / dur) * 100;
      progressFill.style.width = pct + '%';
      playhead.style.left = pct + '%';
      timeDisplay.textContent = `${EdethUI.formatDuration(cur)} / ${EdethUI.formatDuration(dur)}`;

      // Keep orig in sync if drift happens
      if (Math.abs(audioOrig.currentTime - cur) > 0.15) {
        audioOrig.currentTime = cur;
      }
    };

    audioClean.onended = () => {
      playBtn.textContent = '▶';
      progressFill.style.width = '0%';
      playhead.style.left = '0%';
    };

    setTimeout(drawWaveform, 100);

    return card;
  }
};
