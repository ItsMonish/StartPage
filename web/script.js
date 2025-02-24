window.onload = function() {
  document.getElementById("search-query").addEventListener("keypress", function(ev) {
    if (ev.key == "Enter") searchDDG();
  });
}

function searchDDG() {
  let query = document.getElementById("search-query").value;
  window.location.href = "https://duckduckgo.com/?q=" + query;
  document.getElementById("search-query").value = "";
}

function toggleRSS() {
  const transition = document.getElementById('transition');
  const rssFeed = document.getElementById('rss-feed');
  const mainPage = document.querySelector('.container');
  const rssIcon = document.querySelector('.rss-icon');
  // Start the transition effect
  transition.classList.add('transition-active');

  // Delay to let the transition effect finish
  setTimeout(function() {
    // Switch the pages only after the transition completes
    if (mainPage.style.display != 'none') {
      // Show RSS Feed, hide main page
      rssFeed.classList.add('show');
      rssFeed.style.display = 'block';
      mainPage.style.display = 'none';
      rssIcon.innerHTML = '&#x1F50D;';
    } else {
      // Hide RSS Feed, show main page
      rssFeed.classList.remove('show');
      rssFeed.style.display = 'none';
      mainPage.style.display = 'flex';
      rssIcon.innerHTML = '&#x1F4F0;';
    }

    // Remove the transition effect class after switching
    transition.classList.remove('transition-active');
  }, 300); // Match the transition duration
}

