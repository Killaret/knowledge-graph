// Bookmarklet for Knowledge Graph: capture page title, URL and selected text.
// 1. Replace YOUR_DOMAIN with your actual Knowledge Graph domain.
// 2. Copy the value of the BOOKMARKLET constant (without the surrounding quotes)
//    and paste it as the URL of a new browser bookmark.
const BOOKMARKLET = "javascript:(function(){var title=encodeURIComponent(document.title);var url=encodeURIComponent(window.location.href);var text=encodeURIComponent(window.getSelection().toString());window.open('https://YOUR_DOMAIN/import?title='+title+'&url='+url+'&text='+text,'_blank');})();";

export default BOOKMARKLET;
