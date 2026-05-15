// ===== OVN TROUBLESHOOTER - FRONTEND APPLICATION =====

const API_BASE = '/api';

// ===== UTILITY FUNCTIONS =====

async function apiGet(path) {
    try {
        const resp = await fetch(`${API_BASE}${path}`);
        if (!resp.ok) {
            const err = await resp.json().catch(() => ({ error: resp.statusText }));
            throw new Error(err.error || `HTTP ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error(`API GET ${path} failed:`, err);
        throw err;
    }
}

async function apiPost(path, body) {
    try {
        const resp = await fetch(`${API_BASE}${path}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        if (!resp.ok) {
            const err = await resp.json().catch(() => ({ error: resp.statusText }));
            throw new Error(err.error || `HTTP ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error(`API POST ${path} failed:`, err);
        throw err;
    }
}

function showLoading() {
    document.getElementById('loading-overlay').style.display = 'flex';
}

function hideLoading() {
    document.getElementById('loading-overlay').style.display = 'none';
}

async function withLoading(fn) {
    showLoading();
    try {
        return await fn();
    } finally {
        hideLoading();
    }
}

function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function formatJSON(obj) {
    return JSON.stringify(obj, null, 2);
}

function syntaxHighlight(json) {
    json = json.replace(/&/g, '&').replace(/</g, '<').replace(/>/g, '>');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        let cls = 'json-number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'json-key';
            } else {
                cls = 'json-string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'json-boolean';
        } else if (/null/.test(match)) {
            cls = 'json-null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}

function renderJSON(obj) {
    return '<pre class="json-view">' + syntaxHighlight(formatJSON(obj)) + '</pre>';
}

function renderKVDisplay(data) {
    if (!data || typeof data !== 'object') return '';
    let html = '<div class="kv-display">';
    for (const [key, value] of Object.entries(data)) {
        let displayValue = value;
        if (typeof value === 'object') {
            displayValue = formatJSON(value);
        }
        html += `<div class="kv-row"><span class="kv-key">${escapeHtml(key)}</span><span class="kv-value">${escapeHtml(String(displayValue))}</span></div>`;
    }
    html += '</div>';
    return html;
}

function renderExternalIds(externalIds) {
    if (!externalIds || typeof externalIds !== 'object') return '';
    let html = '<div class="external-ids">';
    for (const [key, value] of Object.entries(externalIds)) {
        html += `<span class="ext-id"><span class="key">${escapeHtml(key)}</span>=<span class="val">${escapeHtml(String(value))}</span></span>`;
    }
    html += '</div>';
    return html;
}

function renderTable(headers, rows) {
    if (!rows.length) return '<p class="text-muted">No data available</p>';
    let html = '<table class="data-table"><thead><tr>';
    headers.forEach(h => html += `<th>${escapeHtml(h)}</th>`);
    html += '</tr></thead><tbody>';
    rows.forEach(row => {
        html += '<tr>';
        row.forEach(cell => html += `<td>${escapeHtml(String(cell))}</td>`);
        html += '</tr>';
    });
    html += '</tbody></table>';
    return html;
}

// ===== BREADCRUMB / DRILL-DOWN NAVIGATION =====

let breadcrumbStack = [];

function pushBreadcrumb(label, onClick) {
    breadcrumbStack.push({ label, onClick });
    renderBreadcrumbs();
}

function popBreadcrumb() {
    breadcrumbStack.pop();
    renderBreadcrumbs();
}

function clearBreadcrumbs() {
    breadcrumbStack = [];
    renderBreadcrumbs();
}

function renderBreadcrumbs() {
    const el = document.getElementById('breadcrumb-bar');
    if (!el) return;
    
    if (!breadcrumbStack.length) {
        el.style.display = 'none';
        return;
    }
    
    el.style.display = 'flex';
    let html = '<span class="breadcrumb-item" onclick="clearBreadcrumbs();navigateTo(\'dashboard\')">🏠 Dashboard</span>';
    breadcrumbStack.forEach((b, i) => {
        html += '<span class="breadcrumb-sep">›</span>';
        if (i < breadcrumbStack.length - 1) {
            html += `<span class="breadcrumb-item" onclick="navigateBreadcrumb(${i})">${escapeHtml(b.label)}</span>`;
        } else {
            html += `<span class="breadcrumb-item active">${escapeHtml(b.label)}</span>`;
        }
    });
    el.innerHTML = html;
}

function navigateBreadcrumb(index) {
    // Navigate to the clicked breadcrumb level
    // This is handled by individual view functions
}

// ===== NAVIGATION =====

document.querySelectorAll('.nav-item[data-page]').forEach(item => {
    item.addEventListener('click', () => {
        const page = item.dataset.page;
        navigateTo(page);
    });
});

// Track which pages have been loaded to avoid duplicate loads
const loadedPages = new Set();

function navigateTo(page) {
    clearBreadcrumbs();
    
    // Close any open detail views
    const routerList = document.getElementById('routers-list');
    const routerDetail = document.getElementById('router-detail');
    if (routerList && routerDetail) {
        routerList.style.display = '';
        routerDetail.style.display = 'none';
    }
    const switchList = document.getElementById('switches-list');
    const switchDetail = document.getElementById('switch-detail');
    if (switchList && switchDetail) {
        switchList.style.display = '';
        switchDetail.style.display = 'none';
    }
    const portListContainer = document.getElementById('ports-list-container');
    const portDetail = document.getElementById('port-detail');
    if (portListContainer && portDetail) {
        portListContainer.style.display = '';
        portDetail.style.display = 'none';
    }
    const chassisList = document.getElementById('chassis-list');
    const chassisDetail = document.getElementById('chassis-detail');
    if (chassisList && chassisDetail) {
        chassisList.style.display = '';
        chassisDetail.style.display = 'none';
    }
    
    document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    
    const navItem = document.querySelector(`.nav-item[data-page="${page}"]`);
    const pageEl = document.getElementById(`page-${page}`);
    
    if (navItem) navItem.classList.add('active');
    if (pageEl) pageEl.classList.add('active');
    
    // Auto-load data for each page on first visit
    if (!loadedPages.has(page)) {
        switch (page) {
            case 'routers':
                loadRouters();
                break;
            case 'switches':
                loadSwitches();
                break;
            case 'ports':
                loadPorts();
                break;
            case 'acls':
                loadACLs();
                break;
            case 'chassis':
                loadChassis();
                break;
            case 'flows':
                loadFlows();
                break;
            case 'topology':
                loadTopology();
                break;
            case 'ovs':
                loadOVS();
                break;
        }
        loadedPages.add(page);
    }
}

// ===== DRILL-DOWN: SHOW PORT DETAIL =====

async function showPortDetail(portName, portType) {
    // Navigate to ports page
    navigateTo('ports');
    
    // Set breadcrumb
    pushBreadcrumb(`${portType} port: ${portName}`, null);
    
    const detailEl = document.getElementById('port-detail');
    const listContainer = document.getElementById('ports-list-container');
    if (listContainer) listContainer.style.display = 'none';
    if (!detailEl) return;
    
    detailEl.style.display = 'block';
    document.getElementById('port-detail-title').textContent = `Port: ${portName}`;
    
    try {
        let data;
        if (portType === 'router') {
            data = await apiGet(`/ports/router/${encodeURIComponent(portName)}`);
        } else {
            data = await apiGet(`/ports/switch/${encodeURIComponent(portName)}`);
        }
        document.getElementById('tab-port-overview').innerHTML = renderJSON(data);
    } catch (err) {
        document.getElementById('tab-port-overview').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Show related info tab
    let relatedHtml = '<div class="kv-display">';
    relatedHtml += `<div class="kv-row"><span class="kv-key">Type</span><span class="value">${portType}</span></div>`;
    relatedHtml += `<div class="kv-row"><span class="kv-key">Name</span><span class="value">${escapeHtml(portName)}</span></div>`;
    
    // Add trace link
    relatedHtml += `<div class="kv-row"><span class="kv-key">Actions</span><span class="value"><button class="btn btn-primary" onclick="openTraceForPort('${escapeHtml(portName)}', '${portType}')">🔍 Trace from this port</button></span></div>`;
    relatedHtml += '</div>';
    document.getElementById('tab-port-related').innerHTML = relatedHtml;
    
    // Reset tabs
    const tabs = document.querySelector('#port-detail .detail-tabs');
    if (tabs) {
        tabs.querySelectorAll('.tab-btn').forEach((b, i) => b.classList.toggle('active', i === 0));
    }
    document.querySelectorAll('#port-detail .tab-content').forEach((tc, i) => tc.classList.toggle('active', i === 0));
}

function openTraceForPort(portName, portType) {
    navigateTo('tracer');
    if (portType === 'switch') {
        document.getElementById('trace-lsp').value = portName;
    } else {
        document.getElementById('trace-lrp').value = portName;
    }
}

function closePortDetail() {
    popBreadcrumb();
    const detailEl = document.getElementById('port-detail');
    const listContainer = document.getElementById('ports-list-container');
    if (detailEl) detailEl.style.display = 'none';
    if (listContainer) listContainer.style.display = '';
}

// ===== TAB SWITCHING =====

document.querySelectorAll('.detail-tabs').forEach(tabContainer => {
    tabContainer.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const tabId = btn.dataset.tab;
            const parent = tabContainer.closest('.entity-detail') || tabContainer.closest('.page');
            if (!parent) return;
            
            tabContainer.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            
            parent.querySelectorAll('.tab-content').forEach(tc => tc.classList.remove('active'));
            const target = document.getElementById(`tab-${tabId}`);
            if (target) target.classList.add('active');
        });
    });
});

// ===== CONNECTION STATUS =====

async function updateConnectionStatus() {
    const el = document.getElementById('connection-status');
    try {
        el.className = 'status-indicator connecting';
        el.querySelector('.status-text').textContent = 'Connecting...';
        
        const data = await apiGet('/connection');
        
        if (data.connected) {
            el.className = 'status-indicator connected';
            el.querySelector('.status-text').textContent = 'Connected';
        } else {
            el.className = 'status-indicator disconnected';
            el.querySelector('.status-text').textContent = 'Disconnected';
        }
        
        // Update connection details
        const detailsEl = document.getElementById('connection-details');
        if (detailsEl) {
            detailsEl.innerHTML = `
                <div class="detail-item">
                    <span class="label">Controller</span>
                    <span class="value">${escapeHtml(data.controller_host || 'N/A')}</span>
                </div>
                <div class="detail-item">
                    <span class="label">OVN Version</span>
                    <span class="value">${escapeHtml(data.ovn_version || 'N/A')}</span>
                </div>
                <div class="detail-item">
                    <span class="label">OVS Version</span>
                    <span class="value">${escapeHtml(data.ovs_version || 'N/A')}</span>
                </div>
                <div class="detail-item">
                    <span class="label">NB DB</span>
                    <span class="value">${data.nbdb_connected ? '✅ Connected' : '❌ Disconnected'}</span>
                </div>
                <div class="detail-item">
                    <span class="label">SB DB</span>
                    <span class="value">${data.sbdb_connected ? '✅ Connected' : '❌ Disconnected'}</span>
                </div>
            `;
        }
    } catch (err) {
        el.className = 'status-indicator disconnected';
        el.querySelector('.status-text').textContent = 'Error';
        const detailsEl = document.getElementById('connection-details');
        if (detailsEl) {
            detailsEl.innerHTML = `<div class="info-box error">Failed to connect: ${escapeHtml(err.message)}</div>`;
        }
    }
}

// ===== DASHBOARD =====

async function loadDashboard() {
    try {
        const data = await withLoading(() => apiGet('/dashboard'));
        
        document.getElementById('stat-routers').textContent = data.total_routers ?? '-';
        document.getElementById('stat-switches').textContent = data.total_switches ?? '-';
        document.getElementById('stat-ports').textContent = data.total_ports ?? '-';
        document.getElementById('stat-acls').textContent = data.total_acls ?? '-';
        document.getElementById('stat-chassis').textContent = data.total_chassis ?? '-';
        document.getElementById('stat-nat').textContent = data.total_nat_rules ?? '-';
        document.getElementById('stat-routes').textContent = data.total_static_routes ?? '-';
        document.getElementById('stat-lb').textContent = data.total_load_balancers ?? '-';
    } catch (err) {
        console.error('Failed to load dashboard:', err);
    }
    updateConnectionStatus();
}

// ===== ROUTERS =====

async function loadRouters() {
    try {
        const data = await withLoading(() => apiGet('/routers'));
        const listEl = document.getElementById('routers-list');
        
        if (!data || !Array.isArray(data)) {
            listEl.innerHTML = '<div class="info-box info">No routers found or data format unexpected</div>';
            return;
        }
        
        if (!data.length) {
            listEl.innerHTML = '<div class="info-box info">No logical routers found</div>';
            return;
        }
        
        let html = '';
        data.forEach(router => {
            const name = router.name || router.Name || 'unknown';
            const uuid = router.uuid || router.UUID || '';
            html += `
                <div class="entity-card" onclick="showRouterDetail('${escapeHtml(name)}')">
                    <div>
                        <div class="entity-name">🔀 ${escapeHtml(name)}</div>
                        <div class="entity-meta">UUID: ${escapeHtml(uuid)}</div>
                    </div>
                    <div class="entity-badges">
                        <span class="badge info">Router</span>
                    </div>
                </div>
            `;
        });
        listEl.innerHTML = html;
    } catch (err) {
        document.getElementById('routers-list').innerHTML = `<div class="info-box error">Error loading routers: ${escapeHtml(err.message)}</div>`;
    }
}

async function showRouterDetail(name) {
    document.getElementById('routers-list').style.display = 'none';
    const detailEl = document.getElementById('router-detail');
    detailEl.style.display = 'block';
    document.getElementById('router-detail-title').textContent = `Router: ${name}`;
    
    // Load overview
    try {
        const data = await apiGet(`/routers/${encodeURIComponent(name)}`);
        document.getElementById('tab-router-overview').innerHTML = renderJSON(data);
    } catch (err) {
        document.getElementById('tab-router-overview').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Load static routes
    try {
        const data = await apiGet('/routers/' + encodeURIComponent(name) + '/routes');
        if (Array.isArray(data) && data.length) {
            const routes = data.map(r => [r.prefix || r.Prefix, r.nexthop || r.Nexthop, r.policy || r.Policy || 'simple', r.output_port || r.Output_port || '']);
            document.getElementById('tab-router-routes').innerHTML = renderTable(['Prefix', 'Next Hop', 'Policy', 'Output Port'], routes);
        } else {
            document.getElementById('tab-router-routes').innerHTML = '<div class="info-box info">No static routes found</div>';
        }
    } catch (err) {
        document.getElementById('tab-router-routes').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Load NAT rules
    try {
        const data = await apiGet('/routers/' + encodeURIComponent(name) + '/nat');
        if (Array.isArray(data) && data.length) {
            const nats = data.map(n => [n.external_ip || n.External_ip, n.internal_ip || n.Internal_ip, n.type || n.Type, n.protocol || n.Protocol || '']);
            document.getElementById('tab-router-nat').innerHTML = renderTable(['External IP', 'Internal IP', 'Type', 'Protocol'], nats);
        } else {
            document.getElementById('tab-router-nat').innerHTML = '<div class="info-box info">No NAT rules found</div>';
        }
    } catch (err) {
        document.getElementById('tab-router-nat').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Load flows
    try {
        const data = await apiGet('/routers/' + encodeURIComponent(name) + '/flows');
        if (Array.isArray(data) && data.length) {
            const flows = data.map(f => [f.table !== undefined ? f.table : f.Table, f.priority !== undefined ? f.priority : f.Priority, f.match !== undefined ? f.match : f.Match, f.action !== undefined ? f.action : f.Action]);
            document.getElementById('tab-router-flows').innerHTML = renderTable(['Table', 'Priority', 'Match', 'Action'], flows);
        } else {
            document.getElementById('tab-router-flows').innerHTML = '<div class="info-box info">No flows found</div>';
        }
    } catch (err) {
        document.getElementById('tab-router-flows').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Reset to first tab
    const tabs = document.querySelector('#router-detail .detail-tabs');
    if (tabs) {
        tabs.querySelectorAll('.tab-btn').forEach((b, i) => b.classList.toggle('active', i === 0));
    }
    document.querySelectorAll('#router-detail .tab-content').forEach((tc, i) => tc.classList.toggle('active', i === 0));
}

function closeDetail(type) {
    if (type === 'router') {
        document.getElementById('routers-list').style.display = 'flex';
        document.getElementById('router-detail').style.display = 'none';
    } else if (type === 'switch') {
        document.getElementById('switches-list').style.display = 'flex';
        document.getElementById('switch-detail').style.display = 'none';
    }
}

// ===== SWITCHES =====

async function loadSwitches() {
    try {
        const data = await withLoading(() => apiGet('/switches'));
        const listEl = document.getElementById('switches-list');
        
        if (!data || !Array.isArray(data)) {
            listEl.innerHTML = '<div class="info-box info">No switches found or data format unexpected</div>';
            return;
        }
        
        if (!data.length) {
            listEl.innerHTML = '<div class="info-box info">No logical switches found</div>';
            return;
        }
        
        let html = '';
        data.forEach(sw => {
            const name = sw.name || sw.Name || 'unknown';
            const uuid = sw.uuid || sw.UUID || '';
            html += `
                <div class="entity-card" onclick="showSwitchDetail('${escapeHtml(name)}')">
                    <div>
                        <div class="entity-name">🔌 ${escapeHtml(name)}</div>
                        <div class="entity-meta">UUID: ${escapeHtml(uuid)}</div>
                    </div>
                    <div class="entity-badges">
                        <span class="badge info">Switch</span>
                    </div>
                </div>
            `;
        });
        listEl.innerHTML = html;
    } catch (err) {
        document.getElementById('switches-list').innerHTML = `<div class="info-box error">Error loading switches: ${escapeHtml(err.message)}</div>`;
    }
}

async function showSwitchDetail(name) {
    document.getElementById('switches-list').style.display = 'none';
    const detailEl = document.getElementById('switch-detail');
    detailEl.style.display = 'block';
    document.getElementById('switch-detail-title').textContent = `Switch: ${name}`;
    
    // Load overview
    try {
        const data = await apiGet(`/switches/${encodeURIComponent(name)}`);
        document.getElementById('tab-switch-overview').innerHTML = renderJSON(data);
    } catch (err) {
        document.getElementById('tab-switch-overview').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Load ACLs
    try {
        const data = await apiGet('/switches/' + encodeURIComponent(name) + '/acls');
        if (Array.isArray(data) && data.length) {
            const acls = data.map(a => [a.direction || a.Direction, a.priority !== undefined ? a.priority : a.Priority, a.action || a.Action, a.match || a.Match]);
            document.getElementById('tab-switch-acls').innerHTML = renderTable(['Direction', 'Priority', 'Action', 'Match'], acls);
        } else {
            document.getElementById('tab-switch-acls').innerHTML = '<div class="info-box info">No ACLs on this switch</div>';
        }
    } catch (err) {
        document.getElementById('tab-switch-acls').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Load flows
    try {
        const data = await apiGet('/switches/' + encodeURIComponent(name) + '/flows');
        if (Array.isArray(data) && data.length) {
            const flows = data.map(f => [f.table !== undefined ? f.table : f.Table, f.priority !== undefined ? f.priority : f.Priority, f.match !== undefined ? f.match : f.Match, f.action !== undefined ? f.action : f.Action]);
            document.getElementById('tab-switch-flows').innerHTML = renderTable(['Table', 'Priority', 'Match', 'Action'], flows);
        } else {
            document.getElementById('tab-switch-flows').innerHTML = '<div class="info-box info">No flows found</div>';
        }
    } catch (err) {
        document.getElementById('tab-switch-flows').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Reset to first tab
    const tabs = document.querySelector('#switch-detail .detail-tabs');
    if (tabs) {
        tabs.querySelectorAll('.tab-btn').forEach((b, i) => b.classList.toggle('active', i === 0));
    }
    document.querySelectorAll('#switch-detail .tab-content').forEach((tc, i) => tc.classList.toggle('active', i === 0));
}

// ===== PORTS =====

async function loadRouterPorts() {
    document.querySelectorAll('#page-ports .tab-btn').forEach(b => b.classList.remove('active'));
    event.target.classList.add('active');
    
    try {
        const data = await withLoading(() => apiGet('/ports/router'));
        const listEl = document.getElementById('ports-list');
        
        if (!data || !Array.isArray(data) || !data.length) {
            listEl.innerHTML = '<div class="info-box info">No router ports found</div>';
            return;
        }
        
        let html = '';
        data.forEach(port => {
            const name = port.name || port.Name || 'unknown';
            const networks = port.networks || port.Networks || [];
            const mac = port.mac || port.Mac || '';
            const netStr = Array.isArray(networks) ? networks.join(', ') : networks;
            
            html += `
                <div class="entity-card" onclick="showPortDetail('${escapeHtml(name)}', 'router')">
                    <div>
                        <div class="entity-name">🔀 ${escapeHtml(name)} <span class="port-type-badge router">router</span></div>
                        <div class="entity-meta">IP: ${escapeHtml(netStr)} | MAC: ${escapeHtml(mac)}</div>
                    </div>
                </div>
            `;
        });
        listEl.innerHTML = html;
    } catch (err) {
        document.getElementById('ports-list').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
}

async function loadSwitchPorts() {
    document.querySelectorAll('#page-ports .tab-btn').forEach(b => b.classList.remove('active'));
    event.target.classList.add('active');
    
    try {
        const data = await withLoading(() => apiGet('/ports/switch'));
        const listEl = document.getElementById('ports-list');
        
        if (!data || !Array.isArray(data) || !data.length) {
            listEl.innerHTML = '<div class="info-box info">No switch ports found</div>';
            return;
        }
        
        let html = '';
        data.forEach(port => {
            const name = port.name || port.Name || 'unknown';
            const type = port.type || port.Type || 'interface';
            const addresses = port.addresses || port.Addresses || [];
            const addrStr = Array.isArray(addresses) ? addresses.join(', ') : addresses;
            
            html += `
                <div class="entity-card" onclick="showPortDetail('${escapeHtml(name)}', 'switch')">
                    <div>
                        <div class="entity-name">🔌 ${escapeHtml(name)} <span class="port-type-badge ${type}">${escapeHtml(type)}</span></div>
                        <div class="entity-meta">Addresses: ${escapeHtml(addrStr)}</div>
                    </div>
                </div>
            `;
        });
        listEl.innerHTML = html;
    } catch (err) {
        document.getElementById('ports-list').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
}

function loadPorts() {
    loadRouterPorts();
}

// ===== ACLs =====

async function loadACLs() {
    try {
        const data = await withLoading(() => apiGet('/acls'));
        const listEl = document.getElementById('acls-list');
        
        if (!data || !Array.isArray(data) || !data.length) {
            listEl.innerHTML = '<div class="info-box info">No ACLs found</div>';
            return;
        }
        
        // Apply filters
        const directionFilter = document.getElementById('acl-direction-filter').value;
        const actionFilter = document.getElementById('acl-action-filter').value;
        const searchFilter = document.getElementById('acl-search').value.toLowerCase();
        
        let filtered = data;
        if (directionFilter) {
            filtered = filtered.filter(a => (a.direction || a.Direction || '') === directionFilter);
        }
        if (actionFilter) {
            filtered = filtered.filter(a => (a.action || a.Action || '') === actionFilter);
        }
        if (searchFilter) {
            filtered = filtered.filter(a => {
                const match = (a.match || a.Match || '').toLowerCase();
                return match.includes(searchFilter);
            });
        }
        
        if (!filtered.length) {
            listEl.innerHTML = '<div class="info-box info">No ACLs match the current filters</div>';
            return;
        }
        
        let html = '';
        filtered.forEach(acl => {
            const direction = acl.direction || acl.Direction || '';
            const priority = acl.priority !== undefined ? acl.priority : (acl.Priority !== undefined ? acl.Priority : '-');
            const action = acl.action || acl.Action || '';
            const match = acl.match || acl.Match || '';
            
            const actionClass = ['allow', 'accept', 'allow-related'].includes(action) ? 'success' : 
                               ['deny', 'drop', 'reject'].includes(action) ? 'danger' : 'warning';
            
            html += `
                <div class="entity-card">
                    <div style="flex:1">
                        <div class="entity-name">${escapeHtml(match.substring(0, 100))}${match.length > 100 ? '...' : ''}</div>
                        <div class="entity-meta">Priority: ${escapeHtml(String(priority))}</div>
                    </div>
                    <div class="entity-badges">
                        <span class="badge info">${escapeHtml(direction)}</span>
                        <span class="badge ${actionClass}">${escapeHtml(action)}</span>
                    </div>
                </div>
            `;
        });
        listEl.innerHTML = html;
    } catch (err) {
        document.getElementById('acls-list').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
}

// ===== CHASSIS =====

async function loadChassis() {
    try {
        const data = await withLoading(() => apiGet('/chassis'));
        const listEl = document.getElementById('chassis-list');
        
        if (!data || !Array.isArray(data) || !data.length) {
            listEl.innerHTML = '<div class="info-box info">No chassis found</div>';
            return;
        }
        
        let html = '';
        data.forEach(ch => {
            const name = ch.name || ch.Name || 'unknown';
            const hostname = ch.hostname || ch.Hostname || '';
            const arch = ch.arch || ch.Arch || '';
            const encapTypes = ch.encap_types || ch.EncapTypes || [];
            const encapStr = Array.isArray(encapTypes) ? encapTypes.join(', ') : (encapTypes || '');
            
            html += `
                <div class="entity-card" onclick="showChassisDetail('${escapeHtml(name)}')">
                    <div>
                        <div class="entity-name">🖥️ ${escapeHtml(name)}</div>
                        <div class="entity-meta">Hostname: ${escapeHtml(hostname)} ${arch ? '| ' + escapeHtml(arch) : ''}</div>
                    </div>
                    <div class="entity-badges">
                        <span class="badge success">Online</span>
                        ${encapStr ? `<span class="badge info">${escapeHtml(encapStr)}</span>` : ''}
                    </div>
                </div>
            `;
        });
        listEl.innerHTML = html;
    } catch (err) {
        document.getElementById('chassis-list').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
}

async function showChassisDetail(name) {
    document.getElementById('chassis-list').style.display = 'none';
    const detailEl = document.getElementById('chassis-detail');
    detailEl.style.display = 'block';
    document.getElementById('chassis-detail-title').textContent = `Chassis: ${name}`;
    pushBreadcrumb(`Chassis: ${name}`, null);
    
    // Load overview
    try {
        const data = await apiGet('/chassis');
        const chassis = Array.isArray(data) ? data.find(c => (c.name || c.Name) === name) : null;
        if (chassis) {
            document.getElementById('tab-chassis-overview').innerHTML = renderJSON(chassis);
        } else {
            document.getElementById('tab-chassis-overview').innerHTML = '<div class="info-box info">Chassis data not found</div>';
        }
    } catch (err) {
        document.getElementById('tab-chassis-overview').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Load port bindings for this chassis first (needed for hosted resources)
    (async () => {
    try {
        const pbs = await apiGet('/port-bindings');
        let portList = Array.isArray(pbs) ? pbs : [];
        // Filter port bindings that belong to this chassis
        const chassisPorts = portList.filter(pb => {
            const pbChassis = pb.chassis || pb.Chassis;
            return pbChassis === name || (Array.isArray(pbChassis) && pbChassis.includes(name));
        });
        
        if (chassisPorts.length) {
            let html = '<div class="entity-list">';
            chassisPorts.forEach(pb => {
                const portName = pb.name || pb.PortName || pb.port_name || 'unknown';
                const type = pb.type || pb.Type || 'interface';
                const mac = pb.mac || pb.Mac || '';
                const addresses = pb.addresses || pb.Addresses || [];
                const addrStr = Array.isArray(addresses) ? addresses.join(', ') : (addresses || '');
                
                html += `
                    <div class="entity-card" onclick="showPortDetail('${escapeHtml(portName)}', 'switch')">
                        <div>
                            <div class="entity-name">🔗 ${escapeHtml(portName)} <span class="port-type-badge ${type}">${escapeHtml(type)}</span></div>
                            <div class="entity-meta">MAC: ${escapeHtml(mac)} ${addrStr ? '| ' + escapeHtml(addrStr) : ''}</div>
                        </div>
                    </div>
                `;
            });
            html += '</div>';
            document.getElementById('tab-chassis-ports').innerHTML = html;
        } else {
            document.getElementById('tab-chassis-ports').innerHTML = '<div class="info-box info">No port bindings on this chassis</div>';
        }
        
        // Now load hosted resources (routers running on this chassis)
        // In real OVN, chassis binding for a router is via lrp-set-chassis on the router port
        try {
            const routers = await apiGet('/routers');
            const routerPorts = await apiGet('/ports/router');
            let routerList = Array.isArray(routers) ? routers : [];
            let portList = Array.isArray(routerPorts) ? routerPorts : [];
            
            // Build a map: port name -> chassis
            const portChassisMap = {};
            portList.forEach(p => {
                const pName = p.name || p.Name;
                const pChassis = p.chassis || p.Chassis;
                if (pName && pChassis) {
                    portChassisMap[pName] = pChassis;
                }
            });
            
            // Match routers to chassis: a router runs on a chassis if any of its ports has lrp-set-chassis set
            const chassisRouters = routerList.filter(r => {
                const rPorts = r.ports || r.Ports || [];
                return rPorts.some(rp => portChassisMap[rp] === name);
            });
            
            const switches = await apiGet('/switches');
            let switchList = Array.isArray(switches) ? switches : [];
            // Switches don't directly have chassis, but we can infer from port bindings
            const chassisSwitches = switchList.filter(s => {
                const ports = s.ports || s.Ports || [];
                return ports.some(p => chassisPorts.some(cp => (cp.name || cp.PortName || cp.port_name) === p));
            });
            
            let html = '';
            
            if (chassisRouters.length) {
                html += '<h4 style="margin:12px 0 8px;color:var(--text-secondary)">Logical Routers</h4>';
                chassisRouters.forEach(r => {
                    const rName = r.name || r.Name || 'unknown';
                    html += `
                        <div class="drill-down-row">
                            <span class="drill-down-icon">🔀</span>
                            <div class="drill-down-info" onclick="navigateTo('routers')">
                                <div class="drill-down-name">${escapeHtml(rName)}</div>
                                <div class="drill-down-meta">Chassis: ${escapeHtml(name)}</div>
                            </div>
                        </div>
                    `;
                });
            }
            
            if (chassisSwitches.length) {
                html += '<h4 style="margin:12px 0 8px;color:var(--text-secondary)">Logical Switches</h4>';
                chassisSwitches.forEach(s => {
                    const sName = s.name || s.Name || 'unknown';
                    html += `
                        <div class="drill-down-row">
                            <span class="drill-down-icon">🔌</span>
                            <div class="drill-down-info" onclick="navigateTo('switches')">
                                <div class="drill-down-name">${escapeHtml(sName)}</div>
                                <div class="drill-down-meta">Chassis: ${escapeHtml(name)}</div>
                            </div>
                        </div>
                    `;
                });
            }
            
            if (!chassisRouters.length && !chassisSwitches.length) {
                html = '<div class="info-box info">No routers or switches hosted on this chassis</div>';
            }
            
            document.getElementById('tab-chassis-hosted').innerHTML = html;
        } catch (err) {
            document.getElementById('tab-chassis-hosted').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
        }
        
    } catch (err) {
        console.error('showChassisDetail error:', err);
    }
    })();
    
    // Reset tabs
    const tabs = document.querySelector('#chassis-detail .detail-tabs');
    if (tabs) {
        tabs.querySelectorAll('.tab-btn').forEach((b, i) => b.classList.toggle('active', i === 0));
    }
    document.querySelectorAll('#chassis-detail .tab-content').forEach((tc, i) => tc.classList.toggle('active', i === 0));
}

function closeChassisDetail() {
    popBreadcrumb();
    document.getElementById('chassis-list').style.display = 'flex';
    document.getElementById('chassis-detail').style.display = 'none';
}

// ===== FLOWS =====

async function loadFlows() {
    try {
        const data = await withLoading(() => apiGet('/flows'));
        const listEl = document.getElementById('flows-list');
        
        if (!data || !Array.isArray(data) || !data.length) {
            listEl.innerHTML = '<div class="info-box info">No flows found</div>';
            return;
        }
        
        // Apply filters
        const typeFilter = document.getElementById('flow-type-filter').value;
        const searchFilter = document.getElementById('flow-search').value.toLowerCase();
        
        let filtered = data;
        if (typeFilter) {
            filtered = filtered.filter(f => (f.type || f.Type || '') === typeFilter);
        }
        if (searchFilter) {
            filtered = filtered.filter(f => {
                const match = (f.match || f.Match || '').toLowerCase();
                const action = (f.action || f.Action || '').toLowerCase();
                return match.includes(searchFilter) || action.includes(searchFilter);
            });
        }
        
        // Limit to 500 entries for performance
        const display = filtered.slice(0, 500);
        
        let html = '';
        display.forEach(flow => {
            const table = flow.table !== undefined ? flow.table : (flow.Table || '-');
            const priority = flow.priority !== undefined ? flow.priority : (flow.Priority || '-');
            const match = flow.match || flow.Match || '';
            const action = flow.action || flow.Action || '';
            const type = flow.type || flow.Type || '';
            
            const isDrop = /drop|reject/.test(action.toLowerCase());
            
            html += `
                <div class="entity-card" style="cursor:default">
                    <div style="flex:1;overflow:hidden">
                        <div class="entity-name" style="font-family:monospace;font-size:0.8rem">${escapeHtml(match.substring(0, 120))}${match.length > 120 ? '...' : ''}</div>
                        <div class="entity-meta" style="font-family:monospace">Action: ${escapeHtml(action.substring(0, 80))}</div>
                    </div>
                    <div class="entity-badges">
                        <span class="badge info">T:${escapeHtml(String(table))}</span>
                        <span class="badge">P:${escapeHtml(String(priority))}</span>
                        <span class="badge ${isDrop ? 'danger' : ''}">${escapeHtml(type)}</span>
                    </div>
                </div>
            `;
        });
        listEl.innerHTML = html;
        
        if (filtered.length > 500) {
            listEl.innerHTML += `<div class="info-box warning">Showing 500 of ${filtered.length} flows. Use filters to narrow down.</div>`;
        }
    } catch (err) {
        document.getElementById('flows-list').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
}

// ===== PACKET TRACER =====

async function populateTraceSelectors() {
    // Load routers
    try {
        const routers = await apiGet('/autocomplete/routers');
        const routerSelect = document.getElementById('trace-router');
        routerSelect.innerHTML = '<option value="">Select a router...</option>';
        if (Array.isArray(routers)) {
            routers.forEach(name => {
                routerSelect.innerHTML += `<option value="${escapeHtml(name)}">${escapeHtml(name)}</option>`;
            });
        }
    } catch (e) { console.warn('Failed to load routers:', e); }
    
    // Load switches
    try {
        const switches = await apiGet('/autocomplete/switches');
        const switchSelect = document.getElementById('trace-switch');
        switchSelect.innerHTML = '<option value="">Select a switch...</option>';
        if (Array.isArray(switches)) {
            switches.forEach(name => {
                switchSelect.innerHTML += `<option value="${escapeHtml(name)}">${escapeHtml(name)}</option>`;
            });
        }
    } catch (e) { console.warn('Failed to load switches:', e); }
    
    // Load ports
    try {
        const ports = await apiGet('/autocomplete/ports');
        const lrpSelect = document.getElementById('trace-lrp');
        const lspSelect = document.getElementById('trace-lsp');
        
        lrpSelect.innerHTML = '<option value="">Select a router port...</option>';
        if (Array.isArray(ports.router_ports)) {
            ports.router_ports.forEach(name => {
                lrpSelect.innerHTML += `<option value="${escapeHtml(name)}">${escapeHtml(name)}</option>`;
            });
        }
        
        lspSelect.innerHTML = '<option value="">Select a switch port...</option>';
        if (Array.isArray(ports.switch_ports)) {
            ports.switch_ports.forEach(name => {
                lspSelect.innerHTML += `<option value="${escapeHtml(name)}">${escapeHtml(name)}</option>`;
            });
        }
    } catch (e) { console.warn('Failed to load ports:', e); }
}

function toggleTraceMode() {
    const mode = document.getElementById('trace-mode').value;
    document.getElementById('trace-router-fields').style.display = mode === 'router' ? 'block' : 'none';
    document.getElementById('trace-switch-fields').style.display = mode === 'switch' ? 'block' : 'none';
}

function buildPacketSpec() {
    const custom = document.getElementById('pkt-custom').value.trim();
    if (custom) return custom;
    
    const protocol = document.getElementById('pkt-protocol').value;
    const srcIp = document.getElementById('pkt-src-ip').value.trim();
    const dstIp = document.getElementById('pkt-dst-ip').value.trim();
    const l4Proto = document.getElementById('pkt-l4-protocol').value;
    const srcPort = document.getElementById('pkt-src-port').value;
    const dstPort = document.getElementById('pkt-dst-port').value;
    
    let spec = '';
    if (protocol === 'arp') {
        spec = 'arp[source_mac=aa:bb:cc:dd:ee:ff:00, target_mac=aa:bb:cc:dd:ee:ff:01, source_ip=' + (srcIp || '0.0.0.0') + ', target_ip=' + (dstIp || '0.0.0.0') + ']';
    } else {
        const ipField = protocol === 'ip4' ? 'ip4' : 'ip6';
        spec = `${ipField}.src==${srcIp || '0.0.0.0'} && ${ipField}.dst==${dstIp || '0.0.0.0'}`;
        
        if (l4Proto) {
            if (l4Proto === 'icmp') {
                spec += ` && icmp4[type=8, code=0]`;
            } else {
                spec += ` && ${l4Proto}.src==${srcPort || '0'} && ${l4Proto}.dst==${dstPort || '0'}`;
            }
        }
    }
    
    return spec;
}

async function runTrace() {
    const mode = document.getElementById('trace-mode').value;
    const resultEl = document.getElementById('trace-result');
    
    const pkt = buildPacketSpec();
    if (!pkt) {
        resultEl.innerHTML = '<div class="info-box error">Please specify a packet</div>';
        return;
    }
    
    const checkRod = document.getElementById('trace-rod').checked;
    
    let body = { pkt, check_rod: checkRod };
    
    if (mode === 'router') {
        const router = document.getElementById('trace-router').value;
        const lrp = document.getElementById('trace-lrp').value;
        const lsp = document.getElementById('trace-lsp').value;
        
        if (!router || !lsp) {
            resultEl.innerHTML = '<div class="info-box error">Please select a router and destination switch port</div>';
            return;
        }
        
        body.router = router;
        body.lrp = lrp;
        body.lsp = lsp;
    } else {
        const sw = document.getElementById('trace-switch').value;
        const lsp = document.getElementById('trace-lsp').value;
        
        if (!sw || !lsp) {
            resultEl.innerHTML = '<div class="info-box error">Please select a switch and destination switch port</div>';
            return;
        }
        
        body.switch = sw;
        body.lsp = lsp;
    }
    
    resultEl.innerHTML = '<div class="result-placeholder"><div class="spinner"></div><p>Running trace...</p></div>';
    
    try {
        const result = await apiPost('/trace', body);
        
        if (!result.success) {
            resultEl.innerHTML = `<div class="info-box error">Trace failed: ${escapeHtml(result.error || 'Unknown error')}</div>`;
            return;
        }
        
        // Build result HTML
        let html = '';
        
        // Verdict
        const verdict = result.verdict || '';
        if (verdict === 'success') {
            html += '<div class="trace-verdict success">✅ Packet delivered successfully</div>';
        } else if (verdict === 'dropped') {
            html += '<div class="trace-verdict dropped">❌ Packet was dropped</div>';
        } else {
            html += '<div class="trace-verdict">Trace completed</div>';
        }
        
        // Packet info
        html += '<div class="info-box info">Packet: ' + escapeHtml(pkt) + '</div>';
        
        // Flow trace
        if (result.flows && result.flows.length) {
            html += '<h4 style="margin:16px 0 8px;color:var(--text-secondary)">Flow Trace</h4>';
            result.flows.forEach(flow => {
                const isDrop = /drop|reject/.test(flow.action.toLowerCase());
                html += `<div class="trace-flow ${isDrop ? 'drop' : 'highlight'}">${escapeHtml(flow.action || flow.table || '')}</div>`;
            });
        }
        
        // Raw output
        if (result.output) {
            html += '<h4 style="margin:16px 0 8px;color:var(--text-secondary)">Raw Output</h4>';
            html += '<div class="command-output">' + escapeHtml(result.output) + '</div>';
        }
        
        // ROD result
        if (result.radius_of_darkness) {
            const rod = result.radius_of_darkness;
            const rodPass = rod.pass !== undefined ? rod.pass : (rod.Pass || false);
            html += `
                <div class="rod-result">
                    <h4>🔴 Radius of Darkness</h4>
                    <div class="kv-display">
                        <div class="kv-row"><span class="kv-key">Pass</span><span class="value">${rodPass ? '✅ PASS' : '❌ FAIL'}</span></div>
                        <div class="kv-row"><span class="kv-key">Loss</span><span class="value">${rod.loss !== undefined ? rod.loss : rod.Loss}</span></div>
                        <div class="kv-row"><span class="kv-key">Margin</span><span class="value">${rod.margin !== undefined ? rod.margin : rod.Margin}</span></div>
                        <div class="kv-row"><span class="kv-key">Max Pass Loss</span><span class="value">${rod.max_pass_loss !== undefined ? rod.max_pass_loss : rod.Max_pass_loss}</span></div>
                    </div>
                </div>
            `;
        }
        
        resultEl.innerHTML = html;
    } catch (err) {
        resultEl.innerHTML = `<div class="info-box error">Trace error: ${escapeHtml(err.message)}</div>`;
    }
}

// ===== TOPOLOGY =====

async function loadTopology() {
    const outputEl = document.getElementById('topology-output');
    try {
        const data = await withLoading(() => apiGet('/topology/ovn'));
        outputEl.textContent = data.output || 'No topology data available';
    } catch (err) {
        outputEl.textContent = `Error loading topology: ${err.message}`;
    }
}

// ===== OVS =====

async function loadOVS() {
    // OVS Show
    try {
        const data = await apiGet('/ovs/show');
        document.getElementById('tab-ovs-show').innerHTML = `<div class="command-output">${escapeHtml(data.output || 'No OVS data')}</div>`;
    } catch (err) {
        document.getElementById('tab-ovs-show').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Bridges
    try {
        const data = await apiGet('/ovs/bridges');
        if (Array.isArray(data) && data.length) {
            let html = '<div class="entity-list">';
            data.forEach(bridge => {
                const name = bridge.name || bridge.Name || 'unknown';
                const type = bridge.type || bridge.Type || 'internal';
                html += `
                    <div class="entity-card">
                        <div>
                            <div class="entity-name">🌉 ${escapeHtml(name)}</div>
                            <div class="entity-meta">Type: ${escapeHtml(type)}</div>
                        </div>
                    </div>
                `;
            });
            html += '</div>';
            document.getElementById('tab-ovs-bridges').innerHTML = html;
        } else {
            document.getElementById('tab-ovs-bridges').innerHTML = '<div class="info-box info">No OVS bridges found</div>';
        }
    } catch (err) {
        document.getElementById('tab-ovs-bridges').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
    
    // Interfaces
    try {
        const data = await apiGet('/ovs/interfaces');
        if (Array.isArray(data) && data.length) {
            let html = '<div class="entity-list">';
            data.forEach(iface => {
                const name = iface.name || iface.Name || 'unknown';
                const type = iface.type || iface.Type || 'unknown';
                html += `
                    <div class="entity-card">
                        <div>
                            <div class="entity-name">🔗 ${escapeHtml(name)}</div>
                            <div class="entity-meta">Type: ${escapeHtml(type)}</div>
                        </div>
                    </div>
                `;
            });
            html += '</div>';
            document.getElementById('tab-ovs-interfaces').innerHTML = html;
        } else {
            document.getElementById('tab-ovs-interfaces').innerHTML = '<div class="info-box info">No OVS interfaces found</div>';
        }
    } catch (err) {
        document.getElementById('tab-ovs-interfaces').innerHTML = `<div class="info-box error">${escapeHtml(err.message)}</div>`;
    }
}

// ===== INITIALIZATION =====

async function init() {
    showLoading();
    
    try {
        // Load dashboard data
        loadDashboard();
        
        // Load trace selectors
        populateTraceSelectors();
    } finally {
        hideLoading();
    }
}

// Start the app
init();