document.addEventListener('DOMContentLoaded', function () {
  var btn = document.getElementById('downloadBtn');
  if (!btn) return;
  btn.addEventListener('click', function () {
    var original = btn.textContent;
    btn.textContent = 'Iniciando download...';
    setTimeout(function () { btn.textContent = original; }, 2000);
  });

  // Optional: show connection hint if preview fails
  var media = document.querySelector('video, audio, img');
  if (media) {
    media.addEventListener('error', function () {
      console.warn('preview falhou, mas download ainda disponível');
    });
  }
});
