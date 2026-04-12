(function () {
  'use strict';

  // ── Theme Management ───────────────────────────────────────────────────────
  window.toggleTheme = function () {
    const html = document.documentElement;
    const isDark = html.classList.contains('dark');
    const next = isDark ? 'light' : 'dark';
    
    // Explicitly set BOTH classes to ensure no weird state
    html.classList.remove('dark', 'light');
    html.classList.add(next);
    localStorage.setItem('theme', next);
    
    console.log('Theme toggled to:', next);

    // Re-init icons to swap sun/moon properly
    setTimeout(() => {
        if (window.lucide) window.lucide.createIcons();
    }, 50);
  };

  // ── Sidebar Management ─────────────────────────────────────────────────────
  window.toggleSidebar = function () {
    const html = document.documentElement;
    const isCollapsed = html.classList.toggle('sidebar-collapsed');
    localStorage.setItem('sidebar-state', isCollapsed ? 'collapsed' : 'expanded');
    
    console.log('Sidebar state:', isCollapsed ? 'collapsed' : 'expanded');

    // Smooth icons refresh
    if (window.lucide) window.lucide.createIcons();
  };

  window.openMobileSidebar = function () {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.getElementById('sidebar-overlay');
    if (sidebar && overlay) {
      sidebar.classList.remove('-translate-x-full');
      overlay.classList.remove('hidden');
      document.body.classList.add('overflow-hidden');
    }
  };

  window.closeMobileSidebar = function () {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.getElementById('sidebar-overlay');
    if (sidebar && overlay) {
      sidebar.classList.add('-translate-x-full');
      overlay.classList.add('hidden');
      document.body.classList.remove('overflow-hidden');
    }
  };

  // ── User Menu & Dropdowns ──────────────────────────────────────────────────
  window.toggleUserMenu = function () {
    const dropdown = document.getElementById('user-menu-dropdown');
    if (dropdown) {
      dropdown.classList.toggle('hidden');
    }
  };

  window.toggleDropdown = function (id) {
    const dropdown = document.getElementById(id);
    const chevron = document.querySelector(`[data-dropdown-chevron="${id}"]`);
    if (dropdown) {
      dropdown.classList.toggle('hidden');
      if (chevron) {
        chevron.classList.toggle('rotate-90');
      }
    }
  };

  // ── Global Event Delegation ───────────────────────────────────────────────
  document.addEventListener('click', (e) => {
    // Close user menu when clicking outside
    const userMenu = document.getElementById('user-menu-dropdown');
    const userMenuButton = document.getElementById('user-menu-button');
    if (userMenu && userMenuButton && !userMenuButton.contains(e.target) && !userMenu.contains(e.target)) {
      userMenu.classList.add('hidden');
    }
    
    // Close mobile sidebar when clicking overlay
    const overlay = document.getElementById('sidebar-overlay');
    if (overlay && e.target === overlay) {
      window.closeMobileSidebar();
    }
  });

  // ── Initialization ────────────────────────────────────────────────────────
  const init = () => {
    console.log('Initializing APP UI...');
    
    // Restore theme from localStorage or system preference
    const savedTheme = localStorage.getItem('theme');
    const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    const theme = savedTheme || systemTheme;
    
    document.documentElement.classList.remove('dark', 'light');
    document.documentElement.classList.add(theme);

    // Restore sidebar state
    const state = localStorage.getItem('sidebar-state');
    if (state === 'collapsed') {
      document.documentElement.classList.add('sidebar-collapsed');
    } else {
      document.documentElement.classList.remove('sidebar-collapsed');
    }

    if (window.lucide) {
      window.lucide.createIcons();
    }
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

})();
