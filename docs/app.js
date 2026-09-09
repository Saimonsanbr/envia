// envia docs - carrega markdowns e renderiza com marked
const nav = document.getElementById('nav');
const content = document.getElementById('content');
const themeToggle = document.getElementById('theme-toggle');
const menuBtn = document.getElementById('menu-btn');
const sidebar = document.getElementById('sidebar');
const overlay = document.getElementById('overlay');

// Tema
function getTheme() {
  return localStorage.getItem('envia-theme') || (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark');
}
function setTheme(t) {
  document.documentElement.setAttribute('data-theme', t);
  localStorage.setItem('envia-theme', t);
  themeToggle.textContent = t === 'dark' ? '☀' : '○';
}
setTheme(getTheme());
themeToggle.addEventListener('click', () => {
  setTheme(getTheme() === 'dark' ? 'light' : 'dark');
});

// Menu mobile
function openMenu() { sidebar.classList.add('open'); overlay.classList.add('open'); }
function closeMenu() { sidebar.classList.remove('open'); overlay.classList.remove('open'); }
menuBtn.addEventListener('click', openMenu);
overlay.addEventListener('click', closeMenu);

// Carrega markdown
async function loadMD(path) {
  content.innerHTML = '<p>Carregando...</p>';
  try {
    const res = await fetch(path);
    if (!res.ok) throw new Error(res.statusText);
    const md = await res.text();
    // marked com highlight simples
    const html = marked.parse(md);
    content.innerHTML = html;
    // Atualiza nav ativo
    nav.querySelectorAll('a').forEach(a => a.classList.toggle('active', a.dataset.md === path));
    // Fecha menu no mobile após clique
    closeMenu();
    // Scroll topo
    window.scrollTo(0, 0);
    // Destaca código (sem lib externa, só estilo)
  } catch (e) {
    content.innerHTML = `<p style="color:#e55">Erro ao carregar ${path}: ${e.message}</p>`;
  }
}

// Navegação
nav.addEventListener('click', (e) => {
  const a = e.target.closest('a[data-md]');
  if (!a) return;
  e.preventDefault();
  const md = a.dataset.md;
  history.pushState(null, '', '#' + md);
  loadMD(md);
});

// Roteamento por hash
function fromHash() {
  const h = location.hash.slice(1);
  if (h && h.endsWith('.md')) return h;
  return 'content/introducao.md';
}
window.addEventListener('popstate', () => loadMD(fromHash()));
// Carga inicial
loadMD(fromHash());

// Marca brand também carrega intro
document.querySelector('.brand a').addEventListener('click', (e) => {
  e.preventDefault();
  history.pushState(null, '', '#content/introducao.md');
  loadMD('content/introducao.md');
});
