
/* Identity & endpoints */
const WS_URL = "ws://localhost:8080/ws";
const API_ROLL = "http://localhost:8080/roll";
const API_PROPERTY_INFO = "http://localhost:8080/property/info"; // POST {room, playerId, index}
const API_PROPERTY_BUY = "http://localhost:8080/property/buy";  // POST {room, playerId, index}
const API_PROPERTY_AUCTION = "http://localhost:8080/property/auction"; // POST {room, playerId, index, bid}


const playerId = sessionStorage.getItem("playerId") || crypto.randomUUID();
const playerName = sessionStorage.getItem("playerName") || "Player-" + playerId.slice(0, 4);
const gameId = sessionStorage.getItem("gameId") || "007";
sessionStorage.setItem("playerId", playerId);
sessionStorage.setItem("playerName", playerName);
sessionStorage.setItem("gameId", gameId);

/* DOM */
const who = document.getElementById('who');
const roomTag = document.getElementById('roomTag');
const countTag = document.getElementById('countTag');
const playersEl = document.getElementById('players');
const logEl = document.getElementById('log');
const rollBtn = document.getElementById('rollBtn');
const leaveBtn = document.getElementById('leaveBtn');
const boardEl = document.getElementById('board');
const tokensEl = document.getElementById('tokens');

who.textContent = `${playerName} (${playerId.slice(0, 6)}…)`;
roomTag.textContent = `Room: ${gameId}`;

/* State */
let ws;
let roster = new Map();               // id -> {id,name}
let positions = Object.create(null);  // id -> 0..39
const colors = {};
let lastVersion = 0;
let myTurn = false;

function isNewer(msg) {
    if (typeof msg?.version !== "number") return true;
    if (msg.version <= lastVersion) return false;
    lastVersion = msg.version;
    return true;
}

/* Tiles (GO=0 clockwise) */
const tiles = [
    { name: "GO", type: "corner" },
    { name: "Mediterranean Avenue", band: "#955436", icon: "🏠" },
    { name: "Community Chest", type: "chest", icon: "🧰" },
    { name: "Baltic Avenue", band: "#955436", icon: "🏠" },
    { name: "Income Tax", type: "tax", icon: "💵" },
    { name: "Reading Railroad", type: "rail", icon: "🚂" },
    { name: "Oriental Avenue", band: "#aae0fa", icon: "🏢" },
    { name: "Chance", type: "chance", icon: "❓" },
    { name: "Vermont Avenue", band: "#aae0fa", icon: "🏢" },
    { name: "Connecticut Avenue", band: "#aae0fa", icon: "🏢" },
    { name: "Jail / Just Visiting", type: "corner" },
    { name: "St. Charles Place", band: "#d93a96", icon: "🏘️" },
    { name: "Electric Company", type: "util", icon: "💡" },
    { name: "States Avenue", band: "#d93a96", icon: "🏘️" },
    { name: "Virginia Avenue", band: "#d93a96", icon: "🏘️" },
    { name: "Pennsylvania Railroad", type: "rail", icon: "🚆" },
    { name: "St. James Place", band: "#f7941d", icon: "🏨" },
    { name: "Community Chest", type: "chest", icon: "🧰" },
    { name: "Tennessee Avenue", band: "#f7941d", icon: "🏨" },
    { name: "New York Avenue", band: "#f7941d", icon: "🏨" },
    { name: "Free Parking", type: "corner" },
    { name: "Kentucky Avenue", band: "#ed1b24", icon: "🏬" },
    { name: "Chance", type: "chance", icon: "❓" },
    { name: "Indiana Avenue", band: "#ed1b24", icon: "🏬" },
    { name: "Illinois Avenue", band: "#ed1b24", icon: "🏬" },
    { name: "B. & O. Railroad", type: "rail", icon: "🚄" },
    { name: "Atlantic Avenue", band: "#fef200", icon: "🏢" },
    { name: "Ventnor Avenue", band: "#fef200", icon: "🏢" },
    { name: "Water Works", type: "util", icon: "🚰" },
    { name: "Marvin Gardens", band: "#fef200", icon: "🏢" },
    { name: "Go To Jail", type: "corner" },
    { name: "Pacific Avenue", band: "#1fb25a", icon: "🏢" },
    { name: "North Carolina Avenue", band: "#1fb25a", icon: "🏢" },
    { name: "Community Chest", type: "chest", icon: "🧰" },
    { name: "Pennsylvania Avenue", band: "#1fb25a", icon: "🏢" },
    { name: "Short Line", type: "rail", icon: "🚃" },
    { name: "Chance", type: "chance", icon: "❓" },
    { name: "Park Place", band: "#0072bb", icon: "🏙️" },
    { name: "Luxury Tax", type: "tax", icon: "💎" },
    { name: "Boardwalk", band: "#0072bb", icon: "🏙️" },
];

/* 11x11 grid mapping */
const gridMap = (() => {
    const coords = new Array(40);
    for (let i = 0; i <= 10; i++) coords[i] = { r: 11, c: 11 - i };
    for (let i = 1; i <= 10; i++) coords[10 + i] = { r: 11 - i, c: 1 };
    for (let i = 1; i <= 10; i++) coords[20 + i] = { r: 1, c: i + 1 };
    for (let i = 1; i <= 9; i++)  coords[30 + i] = { r: i + 1, c: 11 };
    return coords;
})();

/* Build board (with diagonal decks in center) */
//     function buildBoard() {
//       boardEl.innerHTML = "";
//       for (let i = 0; i < 40; i++) {
//         const t = tiles[i]; const { r, c } = gridMap[i];
//         const d = document.createElement('div');
//         d.className = "tile" + (t.type === "corner" ? " corner" : "");
//         d.style.gridRow = r; d.style.gridColumn = c;
//         if (t.band) { const b = document.createElement("div"); b.className = "band"; b.style.background = t.band; d.appendChild(b); }
//         const ic = document.createElement("div"); ic.className = "icon"; ic.textContent = t.icon || "⬜";
//         const nm = document.createElement("div"); nm.className = "name"; nm.textContent = shortName(t.name);
//         d.appendChild(ic); d.appendChild(nm);
//         boardEl.appendChild(d);
//       }

//       const center = document.createElement('div');
//       center.className = "center";
//       center.id = "center";
//       center.innerHTML = `
//   <div class="title">MONOPOLY</div>
//   <div class="decks">
//     <!-- Top-left: CHANCE -->
//     <div class="deck chance tl" id="deckChance" aria-label="Draw a Chance card" title="Draw a Chance card">
//       <div class="stack">
//         <div class="card">
//           <div class="label"><span class="q">?</span><span>CHANCE</span></div>
//           <div style="font-size:1.4rem;opacity:.7;">Draw the top card</div>
//         </div>
//         <div class="card"></div>
//         <div class="card"></div>
//       </div>
//     </div>

//     <!-- Bottom-right: MYSTERY (Community Chest) -->
//     <div class="deck chest br" id="deckChest" aria-label="Draw a Mystery (Community Chest) card" title="Draw a Mystery (Community Chest) card">
//       <div class="stack">
//         <div class="card">
//           <div class="label"><span class="q">?</span><span>MYSTERY</span></div>
//           <div style="font-size:1.1rem;opacity:.7;">(Community Chest)</div>
//         </div>
//         <div class="card"></div>
//         <div class="card"></div>
//       </div>
//     </div>
//   </div>

// <div class="cardModal" id="buyModal" role="dialog" aria-modal="true" aria-labelledby="buyTitle">
//   <div class="cardView">
//     <div class="kicker">PROPERTY</div>
//     <h4 id="buyTitle">—</h4>
//     <p id="buyBody"></p>
//     <div style="display:flex; gap:10px; justify-content:center; margin-top:14px;">
//       <button class="btn" id="buyConfirm">Buy</button>
//       <button class="btn red" id="buyCancel">Skip</button>
//     </div>
//   </div>
// </div>

// `;
//       boardEl.appendChild(center);


//       // Deck click handlers (only active on your turn)
//       center.querySelector('#deckChance').addEventListener('click', () => { if (myTurn) requestDraw('chance'); });
//       center.querySelector('#deckChest').addEventListener('click', () => { if (myTurn) requestDraw('chest'); });
//       center.querySelector('#cardClose').addEventListener('click', hideCard);
//       center.querySelector('#cardModal').addEventListener('click', (e) => { if (e.target.id === 'cardModal') hideCard(); });
//       center.querySelector('#buyCancel').addEventListener('click', hideBuy);
//       center.querySelector('#buyModal').addEventListener('click', (e) => { if (e.target.id === 'buyModal') hideBuy(); });

//     }

/* Build board (with diagonal decks in center) */
function buildBoard() {
    boardEl.innerHTML = "";
    for (let i = 0; i < 40; i++) {
        const t = tiles[i]; const { r, c } = gridMap[i];
        const d = document.createElement('div');
        d.className = "tile" + (t.type === "corner" ? " corner" : "");
        d.style.gridRow = r; d.style.gridColumn = c;
        if (t.band) { const b = document.createElement("div"); b.className = "band"; b.style.background = t.band; d.appendChild(b); }
        const ic = document.createElement("div"); ic.className = "icon"; ic.textContent = t.icon || "⬜";
        const nm = document.createElement("div"); nm.className = "name"; nm.textContent = shortName(t.name);
        d.appendChild(ic); d.appendChild(nm);
        boardEl.appendChild(d);
    }

    const center = document.createElement('div');
    center.className = "center";
    center.id = "center";
    center.innerHTML = `
    <div class="title">MONOPOLY</div>
    <div class="decks">
      <!-- Top-left: CHANCE -->
      <div class="deck chance tl" id="deckChance" aria-label="Draw a Chance card" title="Draw a Chance card">
        <div class="stack">
          <div class="card">
            <div class="label"><span class="q">?</span><span>CHANCE</span></div>
            <div style="font-size:1.4rem;opacity:.7;">Draw the top card</div>
          </div>
          <div class="card"></div>
          <div class="card"></div>
        </div>
      </div>

      <!-- Bottom-right: MYSTERY (Community Chest) -->
      <div class="deck chest br" id="deckChest" aria-label="Draw a Mystery (Community Chest) card" title="Draw a Mystery (Community Chest) card">
        <div class="stack">
          <div class="card">
            <div class="label"><span class="q">?</span><span>MYSTERY</span></div>
            <div style="font-size:1.1rem;opacity:.7;">(Community Chest)</div>
          </div>
          <div class="card"></div>
          <div class="card"></div>
        </div>
      </div>
    </div>

    <!-- Chance/Mystery card reveal modal -->
    <div class="cardModal" id="cardModal" role="dialog" aria-modal="true" aria-labelledby="cardTitle">
      <div class="cardView">
        <div class="kicker" id="cardDeckKicker">CARD</div>
        <h4 id="cardTitle">—</h4>
        <p id="cardBody"></p>
        <button class="btn closeBtn" id="cardClose">OK</button>
      </div>
    </div>

    <!-- Buy property modal -->
    <div class="cardModal" id="buyModal" role="dialog" aria-modal="true" aria-labelledby="buyTitle">
      <div class="cardView">
        <div class="kicker">PROPERTY</div>
        <h4 id="buyTitle">—</h4>
        <p id="buyBody"></p>
        <div style="display:flex; gap:10px; justify-content:center; margin-top:14px;">
          <button class="btn" id="buyConfirm">Buy</button>
          <button class="btn red" id="buyCancel">Skip</button>
        </div>
      </div>
    </div>
  `;

    // 1) Append to DOM first
    boardEl.appendChild(center);

    // 2) Now wire listeners safely
    const deckChance = document.getElementById('deckChance');
    const deckChest = document.getElementById('deckChest');
    const cardClose = document.getElementById('cardClose');
    const cardModal = document.getElementById('cardModal');
    const buyCancel = document.getElementById('buyCancel');
    const buyModal = document.getElementById('buyModal');

    deckChance?.addEventListener('click', () => { if (myTurn) requestDraw('chance'); });
    deckChest?.addEventListener('click', () => { if (myTurn) requestDraw('chest'); });

    cardClose?.addEventListener('click', hideCard);
    cardModal?.addEventListener('click', (e) => { if (e.target.id === 'cardModal') hideCard(); });

    buyCancel?.addEventListener('click', hideBuy);
    buyModal?.addEventListener('click', (e) => { if (e.target.id === 'buyModal') hideBuy(); });
}

function shortName(n) {
    return n.replace("Avenue", "Ave").replace("Railroad", "RR").replace("Pennsylvania", "Penn")
        .replace("Carolina", "Car.").replace("Community Chest", "Chest")
        .replace("Electric Company", "Electric Co.").replace("Water Works", "Water");
}

function isPurchasable(index) {
    const t = tiles[index];
    if (!t) return false;
    // Color properties have a band; rail & util are also purchasable.
    if (t.band) return true;
    if (t.type === 'rail' || t.type === 'util') return true;
    return false; // corners, taxes, chance, chest, etc.
}


function showBuy(name, price, index) {
    const m = document.getElementById('buyModal');
    document.getElementById('buyTitle').textContent = `Buy ${name}?`;
    document.getElementById('buyBody').textContent = `Price: $${price}`;
    const btn = document.getElementById('buyConfirm');
    btn.disabled = false;
    btn.onclick = async () => {
        btn.disabled = true;
        try {
            const res = await fetch(API_PROPERTY_BUY, {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ room: gameId, playerId, index })
            });
            const payload = await (async () => { try { return await res.json(); } catch { return null; } })();
            if (!res.ok) {
                logLine(`Buy failed: ${res.status} ${res.statusText}${payload?.error ? " — " + payload.error : ""}`);
            } else {
                logLine(`Purchased ${name} for $${price}.`);
            }
        } catch (e) {
            logLine(`Error calling /property/buy: ${e}`);
        }
        hideBuy();
    };
    m.classList.add('show');
}
function hideBuy() { document.getElementById('buyModal').classList.remove('show'); }

document.getElementById('buyCancel')?.addEventListener('click', hideBuy);
document.getElementById('buyModal')?.addEventListener('click', (e) => { if (e.target.id === 'buyModal') hideBuy(); });




/* Card modal helpers */
function showCard(deck, title, body) {
    const modal = document.getElementById('cardModal');
    document.getElementById('cardDeckKicker').textContent = deck === 'chance' ? '? CHANCE' : '? MYSTERY';
    document.getElementById('cardTitle').textContent = title || 'Card';
    document.getElementById('cardBody').textContent = body || '';
    modal.classList.add('show');
}
function hideCard() { document.getElementById('cardModal').classList.remove('show'); }
function requestDraw(deck) { send({ type: "draw", deck, room: gameId, playerId }); logLine(`Requested a ${deck === 'chance' ? 'Chance' : 'Mystery'} card…`); }

/* Tokens */
function ensureToken(pid, name) {
    let el = document.getElementById("t_" + pid);
    if (!el) {
        el = document.createElement("div");
        el.id = "t_" + pid; el.className = "token";
        el.dataset.initial = (name || "?").trim()[0]?.toUpperCase() || "?";
        colors[pid] = colors[pid] || pickColor(pid);
        el.style.setProperty("--token", colors[pid]);
        tokensEl.appendChild(el);
    }
    return el;
}
function moveToken(pid, toIndex) {
    const pos = Math.max(0, Math.min(39, toIndex | 0));
    positions[pid] = pos;
    const { r, c } = gridMap[pos];
    const el = ensureToken(pid, roster.get(pid)?.name || "");
    const { dx, dy } = tinyOffset(pid);
    el.style.gridRow = r; el.style.gridColumn = c;
    el.style.setProperty("--tx", dx + "px"); el.style.setProperty("--ty", dy + "px");
}
function animateMove(pid, from, to, onDone) {
    let steps = ((to - from) + 40) % 40;
    if (steps === 0 && from !== to) steps = 40;
    let cur = from, i = 0;
    const hop = () => {
        cur = (cur + 1) % 40; moveToken(pid, cur);
        if (++i < steps) setTimeout(hop, 80);
        else if (typeof onDone === 'function') onDone(cur);
    };
    hop();
}
function pickColor(id) { const p = ["#1f6feb", "#d64545", "#1fb25a", "#7c3aed", "#eab308", "#ef4444", "#06b6d4", "#f97316", "#3b82f6", "#16a34a"]; let h = 0; for (let i = 0; i < id.length; i++) h = (h * 33 + id.charCodeAt(i)) >>> 0; return p[h % p.length]; }
function tinyOffset(id) { let h = 0; for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) >>> 0; return { dx: ((h & 7) - 3) * 3, dy: (((h >> 3) & 7) - 3) * 3 }; }

/* Logging & roster */
function t() { return new Date().toLocaleTimeString(); }
function logLine(text, cls = "sys") { const d = document.createElement('div'); d.className = cls; d.textContent = `[${t()}] ${text}`; logEl.appendChild(d); logEl.scrollTop = logEl.scrollHeight; }
function logInbound(payload) { const d = document.createElement('div'); d.className = 'in'; d.textContent = `[${t()}] ← ${payload}`; logEl.appendChild(d); logEl.scrollTop = logEl.scrollHeight; }
function logOutbound(obj) { const d = document.createElement('div'); d.className = 'out'; d.textContent = `[${t()}] → ${JSON.stringify(obj)}`; logEl.appendChild(d); logEl.scrollTop = logEl.scrollHeight; }
function escapeHtml(s) { return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", " >": "&gt;", "\"": "&quot;", "'": "&#39;" }[c] || c)); }
function renderPlayers(list) {
    roster = new Map(list.map(p => [p.id, p]));
    playersEl.innerHTML = ""; countTag.textContent = `${list.length}/10`;
    list.forEach(p => {
        const el = document.createElement('div'); el.className = 'roster-item';
        const me = p.id === playerId;
        el.innerHTML = `<div><span class="pill ${me ? 'me' : ''}">${me ? 'You' : 'Player'}</span> <strong>${escapeHtml(p.name || '')}</strong></div><div class="id">${escapeHtml(p.id.slice(0, 6))}…</div>`;
        playersEl.appendChild(el);
        if (!(p.id in positions)) moveToken(p.id, 0);
    });
    Array.from(tokensEl.children).forEach(tok => { const pid = tok.id.slice(2); if (!roster.has(pid)) tok.remove(); });
}

/* WebSocket */
function send(obj) { if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify(obj)); logOutbound(obj); } }

function connect() {
    ws = new WebSocket(WS_URL);
    ws.onopen = () => {
        logLine("Connected.");
        send({ type: "resume", playerId, name: playerName, room: gameId });
        send({ type: "subscribeLogs", room: gameId });
        send({ type: "who", room: gameId });
        send({ type: "sync", room: gameId });
    };
    ws.onmessage = ev => {
        const raw = ev.data; logInbound(typeof raw === "string" ? raw : "[binary]");
        let msg; try { msg = JSON.parse(raw) } catch { return; }

        switch (msg.type) {
            case "players": renderPlayers(msg.list || []); break;

            case "playerJoined":
                if (msg.player) { roster.set(msg.player.id, msg.player); renderPlayers([...roster.values()]); }
                break;

            case "playerLeft":
                if (msg.player) { roster.delete(msg.player.id); renderPlayers([...roster.values()]); }
                break;

            case "yourTurn":
                myTurn = !!msg.canRoll;
                rollBtn.disabled = !myTurn;
                if (myTurn) logLine("It's your turn!");
                break;

            case "event":
            case "log":
            case "serverLog":
                if (msg.text) logLine(msg.text);
                break;

            case "move": {
                if (!isNewer(msg)) return;
                const pid = msg.playerId; if (!pid) break;
                const from = Number(positions[pid] ?? 0);
                const to = Number(msg.to ?? from);
                animateMove(pid, from, ((to % 40) + 40) % 40, (landed) => {
                    if (pid === playerId) maybeAutoDraw(landed);
                });
                break;
            }

            case "state": {
                if (!isNewer(msg)) return;
                if (msg.positions && typeof msg.positions === "object") {
                    Object.entries(msg.positions).forEach(([pid, idx]) => {
                        const cur = Number(positions[pid] ?? 0);
                        const target = Number(idx ?? cur);
                        if (cur !== target) {
                            animateMove(pid, cur, ((target % 40) + 40) % 40, (landed) => {
                                if (pid === playerId) maybeAutoDraw(landed);
                            });
                        }
                    });
                }
                if (msg.turn) {
                    myTurn = (msg.turn === playerId);
                    rollBtn.disabled = !myTurn;
                }
                break;
            }

            case "drawPrompt": {
                if (!msg.deck) break;
                requestDraw(String(msg.deck).toLowerCase() === 'chance' ? 'chance' : 'chest');
                break;
            }

            case "card": {
                if (!msg.deck) break;
                const deck = String(msg.deck).toLowerCase() === 'chance' ? 'chance' : 'chest';
                const title = msg.title || (deck === 'chance' ? 'Advance to GO' : 'Bank error in your favor');
                const body = msg.text || '';
                showCard(deck, title, body);
                break;
            }

            default:
                logLine(`JSON: ${JSON.stringify(msg, null, 2)}`);
        }
    };
    ws.onclose = () => { logLine("Disconnected."); rollBtn.disabled = true; myTurn = false; retry(); };
    ws.onerror = () => { logLine("WebSocket error."); };
}
let retries = 0; function retry() { const d = Math.min(1000 * Math.pow(2, retries++), 8000); setTimeout(connect, d); }
setInterval(() => { send({ type: "ping", t: Date.now(), room: gameId }); }, 25000);

/* Auto draw when landing on chance/chest, only if it's your turn */
function maybeAutoDraw(landedIndex) {
    const t = tiles[landedIndex]; if (!t || !myTurn) return;
    if (t.type === 'chance') requestDraw('chance');
    else if (t.type === 'chest') requestDraw('chest');
}

/* Roll */
rollBtn.addEventListener('click', async () => {
    rollBtn.disabled = true;
    try {
        const res = await fetch(API_ROLL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ playerId, room: gameId, name: playerName })
        });
        const payload = await (async () => { try { return await res.json(); } catch { return null; } })();
        if (!res.ok) {
            logLine(`Roll failed: ${res.status} ${res.statusText}${payload?.error ? " — " + payload.error : ""}`);
            send({ type: "sync", room: gameId });
            return;
        }
        if (payload?.dice?.length === 2) {
            logLine(`Roll requested. Server dice: ${payload.dice[0]} & ${payload.dice[1]} (total ${Number(payload.total ?? (payload.dice[0] + payload.dice[1]))}).`);
        } else {
            logLine(`Roll requested.`);
        }
    } catch (e) {
        logLine(`Error calling /roll: ${e}`);
        send({ type: "sync", room: gameId });
    }
});

leaveBtn.addEventListener('click', () => {
    try { send({ type: "leave", playerId, room: gameId }); ws && ws.close(1000); } catch { }
    sessionStorage.removeItem("playerId"); sessionStorage.removeItem("playerName"); sessionStorage.removeItem("gameId");
    window.location.href = "index.html";
});

buildBoard(); connect();
