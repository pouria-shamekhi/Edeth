/* ============================================
   Edeth File Browser Component
   ============================================ */

const FileBrowser = {
  currentPath: '',
  selectedFile: '',
  acceptExts: [],
  resolvePromise: null,
  mode: 'file',

  open(options = {}) {
    return new Promise((resolve) => {
      this.resolvePromise = resolve;
      this.acceptExts = (options.accept || []).map(e => e.toLowerCase());
      this.mode = options.mode || 'file';
      this.selectedFile = '';
      this.render(options.title || 'Select File');
      this.loadDrivesAndNavigate();
    });
  },

  close(result = null) {
    const modal = document.querySelector('.file-browser-modal');
    if (modal) modal.remove();
    if (this._escHandler) {
      document.removeEventListener('keydown', this._escHandler);
      this._escHandler = null;
    }
    if (this.resolvePromise) {
      const fn = this.resolvePromise;
      this.resolvePromise = null;
      fn(result);
    }
  },

  render(title) {
    const existing = document.querySelector('.file-browser-modal');
    if (existing) existing.remove();

    const modal = document.createElement('div');
    modal.className = 'file-browser-modal';
    modal.onclick = (e) => { if (e.target === modal) this.close(null); };

    const content = document.createElement('div');
    content.className = 'file-browser-content';

    // Header
    const header = document.createElement('div');
    header.className = 'file-browser-header';
    const h3 = document.createElement('h3');
    h3.textContent = title;
    header.appendChild(h3);

    const closeBtn = document.createElement('button');
    closeBtn.className = 'file-browser-close';
    closeBtn.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>';
    closeBtn.onclick = () => this.close(null);
    header.appendChild(closeBtn);
    content.appendChild(header);

    // Breadcrumb
    const breadcrumb = document.createElement('div');
    breadcrumb.className = 'file-browser-breadcrumb';
    breadcrumb.id = 'fb-breadcrumb';
    content.appendChild(breadcrumb);

    // Body
    const body = document.createElement('div');
    body.className = 'file-browser-body';

    // Sidebar with drives & quick access
    const sidebar = document.createElement('div');
    sidebar.className = 'file-browser-sidebar';
    sidebar.id = 'fb-drives';
    body.appendChild(sidebar);

    // File list
    const list = document.createElement('div');
    list.className = 'file-browser-list';
    list.id = 'fb-files';
    body.appendChild(list);

    content.appendChild(body);

    // Footer
    const footer = document.createElement('div');
    footer.className = 'file-browser-footer';

    const selectedText = document.createElement('div');
    selectedText.className = 'file-browser-selected';
    selectedText.id = 'fb-selected';
    selectedText.textContent = 'No file selected';
    footer.appendChild(selectedText);

    const actions = document.createElement('div');
    actions.className = 'file-browser-actions';

    const cancelBtn = document.createElement('button');
    cancelBtn.className = 'btn btn-secondary';
    cancelBtn.textContent = 'Cancel';
    cancelBtn.onclick = () => this.close(null);
    actions.appendChild(cancelBtn);

    const selectBtn = document.createElement('button');
    selectBtn.className = 'btn btn-primary';
    selectBtn.id = 'fb-select-btn';
    selectBtn.textContent = 'Select';
    selectBtn.disabled = true;
    selectBtn.onclick = () => {
      if (this.selectedFile) this.close(this.selectedFile);
    };
    actions.appendChild(selectBtn);

    footer.appendChild(actions);
    content.appendChild(footer);

    modal.appendChild(content);
    document.getElementById('modal-container').appendChild(modal);

    // ESC key to close
    this._escHandler = (e) => {
      if (e.key === 'Escape') {
        this.close(null);
      }
    };
    document.addEventListener('keydown', this._escHandler);
  },

  async loadDrivesAndNavigate() {
    const sidebar = document.getElementById('fb-drives');
    let startPath = this.currentPath;

    try {
      const res = await fetch('/api/drives');
      const data = await res.json();

      if (sidebar) {
        sidebar.innerHTML = '';

        // Drives header
        const drivesTitle = document.createElement('div');
        drivesTitle.style.cssText = 'font-size:0.75rem;color:var(--text-muted);text-transform:uppercase;letter-spacing:0.05em;padding:4px 12px;margin-bottom:4px;';
        drivesTitle.textContent = 'Drives';
        sidebar.appendChild(drivesTitle);

        (data.drives || []).forEach(drive => {
          const item = document.createElement('div');
          item.className = 'drive-item';
          item.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="12" x2="2" y2="12"/><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/></svg>';
          const span = document.createElement('span');
          span.textContent = drive;
          item.appendChild(span);
          item.onclick = () => this.navigateTo(drive + '\\');
          sidebar.appendChild(item);
        });

        // Quick Access items from server
        if (data.quick && data.quick.length > 0) {
          const quickTitle = document.createElement('div');
          quickTitle.style.cssText = 'font-size:0.75rem;color:var(--text-muted);text-transform:uppercase;letter-spacing:0.05em;padding:4px 12px;margin-top:16px;margin-bottom:4px;';
          quickTitle.textContent = 'Quick Access';
          sidebar.appendChild(quickTitle);

          data.quick.forEach(q => {
            const item = document.createElement('div');
            item.className = 'drive-item';
            item.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>';
            const span = document.createElement('span');
            span.textContent = q.name;
            item.appendChild(span);
            item.onclick = () => this.navigateTo(q.path);
            sidebar.appendChild(item);
          });
        }
      }

      if (!startPath) {
        startPath = data.home || (data.drives && data.drives[0] ? data.drives[0] + '\\' : 'C:\\');
      }
    } catch (err) {
      if (!startPath) startPath = 'C:\\';
    }

    this.navigateTo(startPath);
  },

  async navigateTo(path) {
    const list = document.getElementById('fb-files');
    if (!list) return;

    this.currentPath = path;
    this.selectedFile = '';
    this.updateSelectedDisplay();

    list.innerHTML = '<div class="file-browser-loading">Loading...</div>';
    this.renderBreadcrumb();
    this.updateDriveHighlight();

    try {
      const res = await fetch('/api/browse?path=' + encodeURIComponent(path));
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Failed to load directory');
      }
      const data = await res.json();
      this.currentPath = data.path;
      this.renderBreadcrumb();
      this.renderFileList(data.entries || [], data.path, data.parent);
    } catch (err) {
      list.innerHTML = `<div class="file-browser-empty">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        <span>${err.message}</span>
      </div>`;
    }
  },

  renderBreadcrumb() {
    const bc = document.getElementById('fb-breadcrumb');
    if (!bc) return;
    bc.innerHTML = '';

    const parts = this.currentPath.split(/[\\\/]/).filter(Boolean);
    let builtPath = '';

    parts.forEach((part, index) => {
      if (index > 0) {
        const sep = document.createElement('span');
        sep.className = 'breadcrumb-separator';
        sep.textContent = ' › ';
        bc.appendChild(sep);
      }

      builtPath += (index === 0 ? part : '\\' + part);
      const pathAtThisLevel = builtPath + (index === 0 ? '\\' : '');

      const item = document.createElement('span');
      item.className = 'breadcrumb-item';
      item.textContent = part;
      item.onclick = () => this.navigateTo(pathAtThisLevel);
      bc.appendChild(item);
    });
  },

  updateDriveHighlight() {
    const sidebar = document.getElementById('fb-drives');
    if (!sidebar) return;
    const driveLetter = this.currentPath.substring(0, 2).toUpperCase();
    sidebar.querySelectorAll('.drive-item').forEach(item => {
      const text = item.querySelector('span');
      if (text && text.textContent.toUpperCase() === driveLetter) {
        item.classList.add('active');
      } else {
        item.classList.remove('active');
      }
    });
  },

  renderFileList(entries, basePath, parentPath) {
    const list = document.getElementById('fb-files');
    if (!list) return;
    list.innerHTML = '';

    // Add parent directory item if available
    if (parentPath && parentPath !== basePath) {
      const parentItem = document.createElement('div');
      parentItem.className = 'file-item';
      parentItem.innerHTML = `<div class="file-item-icon folder">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
      </div>`;
      const nameEl = document.createElement('span');
      nameEl.className = 'file-item-name';
      nameEl.textContent = '.. (Parent Directory)';
      nameEl.style.fontStyle = 'italic';
      nameEl.style.color = 'var(--text-muted)';
      parentItem.appendChild(nameEl);
      parentItem.onclick = () => this.navigateTo(parentPath);
      list.appendChild(parentItem);
    }

    if (entries.length === 0) {
      const empty = document.createElement('div');
      empty.className = 'file-browser-empty';
      empty.innerHTML = `<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
        <span>This folder is empty</span>`;
      list.appendChild(empty);
      return;
    }

    const sep = (basePath.endsWith('\\') || basePath.endsWith('/')) ? '' : '\\';

    entries.forEach(entry => {
      // Filter non-directories by extension
      if (!entry.isDir && this.acceptExts.length > 0) {
        const ext = (entry.ext || '').toLowerCase();
        if (!this.acceptExts.includes(ext)) return;
      }

      const item = document.createElement('div');
      item.className = 'file-item';

      const iconDiv = document.createElement('div');
      iconDiv.className = 'file-item-icon ' + this.getFileCategory(entry.name, entry.isDir);
      iconDiv.innerHTML = this.getFileIconSVG(entry.name, entry.isDir);
      item.appendChild(iconDiv);

      const nameEl = document.createElement('span');
      nameEl.className = 'file-item-name';
      nameEl.textContent = entry.name;
      item.appendChild(nameEl);

      const sizeEl = document.createElement('span');
      sizeEl.className = 'file-item-size';
      sizeEl.textContent = entry.isDir ? '' : EdethUI.formatFileSize(entry.size);
      item.appendChild(sizeEl);

      const dateEl = document.createElement('span');
      dateEl.className = 'file-item-date';
      dateEl.textContent = entry.modified ? new Date(entry.modified).toLocaleDateString() : '';
      item.appendChild(dateEl);

      const fullPath = basePath + sep + entry.name;

      if (entry.isDir) {
        item.onclick = () => this.navigateTo(fullPath);
      } else {
        item.onclick = () => {
          list.querySelectorAll('.file-item.selected').forEach(el => el.classList.remove('selected'));
          item.classList.add('selected');
          this.selectedFile = fullPath;
          this.updateSelectedDisplay();
        };
        item.ondblclick = () => {
          this.selectedFile = fullPath;
          this.close(this.selectedFile);
        };
      }

      list.appendChild(item);
    });
  },

  updateSelectedDisplay() {
    const selEl = document.getElementById('fb-selected');
    const selBtn = document.getElementById('fb-select-btn');
    if (selEl) {
      if (this.selectedFile) {
        const fileName = this.selectedFile.split(/[\\\/]/).pop();
        selEl.innerHTML = 'Selected: <strong>' + fileName + '</strong>';
      } else {
        selEl.textContent = 'No file selected';
      }
    }
    if (selBtn) {
      selBtn.disabled = !this.selectedFile;
    }
  },

  getFileCategory(name, isDir) {
    if (isDir) return 'folder';
    const ext = (name.split('.').pop() || '').toLowerCase();
    const audioExts = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma', 'm4a', 'opus', 'srm', 'pcm', 'raw'];
    const videoExts = ['mp4', 'mkv', 'avi', 'mov', 'webm', 'flv', 'wmv', 'mpeg', 'mpg'];
    const subExts = ['srt', 'ass', 'ssa', 'vtt', 'sub'];
    if (audioExts.includes(ext)) return 'audio';
    if (videoExts.includes(ext)) return 'video';
    if (subExts.includes(ext)) return 'subtitle';
    return 'other';
  },

  getFileIconSVG(name, isDir) {
    const cat = this.getFileCategory(name, isDir);
    const svgAttrs = 'width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"';
    switch (cat) {
      case 'folder':
        return `<svg ${svgAttrs}><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>`;
      case 'audio':
        return `<svg ${svgAttrs}><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>`;
      case 'video':
        return `<svg ${svgAttrs}><rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/></svg>`;
      case 'subtitle':
        return `<svg ${svgAttrs}><polyline points="4 7 4 4 20 4 20 7"/><line x1="9" y1="20" x2="15" y2="20"/><line x1="12" y1="4" x2="12" y2="20"/></svg>`;
      default:
        return `<svg ${svgAttrs}><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>`;
    }
  }
};
