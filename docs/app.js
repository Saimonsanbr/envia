// envia docs - carrega markdowns e renderiza com marked
const nav = document.getElementById('nav');
const content = document.getElementById('content');
const themeToggle = document.getElementById('theme-toggle');
const langToggle = document.getElementById('lang-toggle');
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

// Idioma
function getLang() {
  return localStorage.getItem('envia-lang') || (navigator.language.startsWith('en') ? 'en' : 'pt');
}
function setLang(l) {
  localStorage.setItem('envia-lang', l);
  document.documentElement.setAttribute('lang', l === 'en' ? 'en' : 'pt-BR');
  langToggle.textContent = l === 'en' ? 'PT' : 'EN';
  // Atualiza nav data-md para apontar para pasta correta
  nav.querySelectorAll('a[data-md]').forEach(a => {
    const base = a.dataset.md.replace(/^content\/en\//, 'content/').replace(/^content\//, '');
    // base é ex: introducao.md, instalacao.md
    // Mas precisamos do path completo: content/introducao.md vs content/en/introducao.md
    // O dataset original guarda sem prefixo? Vamos guardar o base sem lang
    const raw = a.getAttribute('data-md-raw') || a.dataset.md;
    if (!a.hasAttribute('data-md-raw')) a.setAttribute('data-md-raw', raw.replace(/^content\/en\//, '').replace(/^content\//, ''));
    const rawBase = a.getAttribute('data-md-raw');
    a.dataset.md = (l === 'en' ? 'content/en/' : 'content/') + rawBase;
  });
  // Atualiza brand link
  const brandLink = document.querySelector('.brand a');
  if (brandLink) {
    const rawBrand = brandLink.getAttribute('data-md-raw') || brandLink.dataset.md;
    if (!brandLink.hasAttribute('data-md-raw')) brandLink.setAttribute('data-md-raw', rawBrand.replace(/^content\/en\//, '').replace(/^content\//, ''));
    const rawBaseBrand = brandLink.getAttribute('data-md-raw');
    brandLink.dataset.md = (l === 'en' ? 'content/en/' : 'content/') + rawBaseBrand;
  }
}
setLang(getLang());
langToggle.addEventListener('click', () => {
  const next = getLang() === 'en' ? 'pt' : 'en';
  setLang(next);
  // Recarrega o conteúdo atual no novo idioma
  const h = location.hash.slice(1);
  let base = 'introducao.md';
  if (h && h.endsWith('.md')) {
    base = h.replace(/^content\/en\//, '').replace(/^content\//, '');
  } else {
    // pega do nav ativo
    const active = nav.querySelector('a.active');
    if (active) base = active.getAttribute('data-md-raw') || active.dataset.md.replace(/^content\/en\//, '').replace(/^content\//, '');
  }
  const newPath = (next === 'en' ? 'content/en/' : 'content/') + base;
  history.pushState(null, '', '#' + newPath);
  loadMD(newPath);
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

// Roteamento por hash - respeita idioma
function fromHash() {
  const h = location.hash.slice(1);
  if (h && h.endsWith('.md')) return h;
  const lang = getLang();
  return (lang === 'en' ? 'content/en/introducao.md' : 'content/introducao.md');
}
window.addEventListener('popstate', () => loadMD(fromHash()));
// Carga inicial - garante que nav está no idioma correto
// O setLang já atualizou os data-md, então fromHash agora retorna o correto
loadMD(fromHash());

// Marca brand também carrega intro no idioma atual
document.querySelector('.brand a').addEventListener('click', (e) => {
  e.preventDefault();
  const lang = getLang();
  const path = lang === 'en' ? 'content/en/introducao.md' : 'content/introducao.md';
  history.pushState(null, '', '#' + path);
  loadMD(path);
});
