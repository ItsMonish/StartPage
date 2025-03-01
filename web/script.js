window.onload = function() {
  document.getElementById("search-query").addEventListener("keypress", function(ev) {
    if (ev.key == "Enter") searchDDG();
  });
  renderSources();
  renderRSS();
}

function searchDDG() {
  let query = document.getElementById("search-query").value;
  window.location.href = "https://duckduckgo.com/?q=" + query;
  document.getElementById("search-query").value = "";
}

function newTab(url) {
  window.open(url, '_blank').focus();
}

function renderRSS() {
  fetch("/rss")
    .then(response => {
      if (!response.ok) {
        throw new Error("Error collection RSS feed response");
      }
      return response.json();
    })
    .then(jsonRssFeed => {
      let feedList = document.getElementById("feed-list");
      feedList.innerHTML = "";
      for (let idx = 0; idx < jsonRssFeed.length; idx++) {
        let curObj = jsonRssFeed[idx];
        let newNode = document.createElement("div");
        newNode.classList.add("feed-item");
        let header = document.createElement("h3");
        header.innerText = curObj.title;
        let src = document.createElement("p")
        src.innerHTML = "<strong>Source:</strong> " + curObj.source;
        let pubdate = document.createElement("p")
        pubdate.innerHTML = "<strong>Published:</strong> " + curObj.pubDate;
        newNode.appendChild(header)
        newNode.appendChild(src)
        newNode.appendChild(pubdate)
        newNode.setAttribute("onclick", `newTab("${curObj.link}");`);
        feedList.appendChild(newNode)
      }
    })
    .catch(error => {
      console.log("Error: " + error)
    });
}

function renderSources() {
  fetch("/rss/srcs")
    .then(response => {
      if (!response.ok) {
        throw new Error("Error collection RSS feed response");
      }
      return response.json();
    })
    .then(jsonSources => {
      let catFilter = document.getElementById("category-filter");
      catFilter.innerHTML = "";
      let allOption = document.createElement("option");
      allOption.setAttribute("value", "all");
      allOption.innerText = "All Categories";
      catFilter.appendChild(allOption);

      for (let category of Object.keys(jsonSources)) {
        let catSources = document.createElement("option");
        catSources.setAttribute("value", category);
        catSources.innerText = titleCase(category);
        catFilter.appendChild(catSources);

        for (let source of jsonSources[category]) {
          let sourceNode = document.createElement("option");
          sourceNode.setAttribute("value", `${category}/${source}`);
          sourceNode.innerText = titleCase(category) + " - " + titleCase(source);

          catFilter.appendChild(sourceNode);
        }
      }
    });
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
      //rssFeed.style.display = 'block';
      mainPage.style.display = 'none';
      rssIcon.innerHTML = '&#x1F50D;';
    } else {
      // Hide RSS Feed, show main page
      rssFeed.classList.remove('show');
      //rssFeed.style.display = 'none';
      mainPage.style.display = 'flex';
      rssIcon.innerHTML = '&#x1F4F0;';
    }

    // Remove the transition effect class after switching
    transition.classList.remove('transition-active');
  }, 300); // Match the transition duration
}

function titleCase(s) {
  return s.toLowerCase()
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}
