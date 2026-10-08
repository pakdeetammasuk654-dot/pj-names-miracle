document.addEventListener('click', async (event) => {
  const button = event.target.closest('.article-share-copy, .article-share-native');
  if (!button) return;

  const url = button.dataset.shareUrl;
  const title = button.dataset.shareTitle;
  if (button.classList.contains('article-share-native') && navigator.share) {
    try {
      await navigator.share({ title, url });
    } catch (error) {
      if (error.name !== 'AbortError') console.warn('Unable to share article', error);
    }
    return;
  }

  try {
    await navigator.clipboard.writeText(url);
    const originalLabel = button.textContent;
    button.textContent = 'คัดลอกแล้ว';
    window.setTimeout(() => { button.textContent = originalLabel; }, 1800);
  } catch (error) {
    window.prompt('คัดลอกลิงก์บทความนี้', url);
  }
});
