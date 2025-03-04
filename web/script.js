window.onload = function() {
  document.getElementById("search-query").addEventListener("keypress", function(ev) {
    if (ev.key == "Enter") searchDDG();
  });
  document.getElementById("category-filter").addEventListener("change", function(ev) {
    activateFilter(document.getElementById("category-filter").value);
  });
  renderSources();
  renderRSS();
}

function searchDDG() {
  let query = document.getElementById("search-query").value;
  window.location.href = "https://duckduckgo.com/?q=" + query;
  document.getElementById("search-query").value = "";
}

function activateFilter(val) {
  fetch("/rss/" + val)
    .then(response => {
      if (!response.ok) {
        throw new Error("Error collection RSS feed with filter " + val);
      }
      return response.json();
    })
    .then(jsonRssFeed => {
      renderJsonInList(jsonRssFeed)
    })
    .catch(error => {
      console.log(error)
    });
}

function renderJsonInList(jsonRssFeed) {
  let feedList = document.getElementById("feed-list");
  feedList.innerHTML = "";
  for (let idx = 0; idx < jsonRssFeed.length; idx++) {
    let curObj = jsonRssFeed[idx];
    let newNode = document.createElement("div");
    newNode.classList.add("feed-item");
    let header = document.createElement("h3");
    header.innerText = curObj.title;
    let src = document.createElement("p");
    src.innerHTML = "<strong>Source:</strong> " + curObj.source + " / " + titleCase(curObj.category);
    let pubdate = document.createElement("p");
    pubdate.innerHTML = "<strong>Published:</strong> " + prettyDate(curObj.pubDate);
    let markIcon = document.createElement("button");
    markIcon.innerHTML = `&#x1F441`;
    markIcon.setAttribute("onclick", `event.stopPropagation();markAsRead(this, ${curObj.id}, true);`);
    markIcon.classList.add("right-button")
    newNode.appendChild(header);
    newNode.appendChild(src);
    newNode.appendChild(pubdate);
    newNode.appendChild(markIcon);
    newNode.setAttribute("onclick", `newTab(this,"${curObj.id}" ,"${curObj.link}");`);
    feedList.appendChild(newNode);
  }
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
      document.getElementById("rss-bubble").innerText = jsonRssFeed.length;
      renderJsonInList(jsonRssFeed)
    })
    .catch(error => {
      console.log(error)
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
    })
    .catch(error => {
      console.log(error);
    });
}

function markAsRead(caller, id, btn) {
  fetch(`/rss/${id}/read`)
    .then(response => {
      if (!response.ok) {
        throw new Error("Error updating history");
      }
      if (!btn) {
        caller.remove();
      } else {
        caller.parentElement.remove();
      }
      document.getElementById("rss-bubble").innerText = document.getElementById("rss-bubble").innerText - 1;
      return
    })
    .catch(err => {
      console.log(err)
    })
}

function changePage(page) {
  const pages = document.querySelectorAll('.page');
  pages.forEach(p => {
    p.classList.remove('active-page');
  });

  const selectedPage = document.getElementById(`${page}-page`);
  selectedPage.classList.add('active-page');
}

function newTab(caller, id, url) {
  markAsRead(caller, id, false);
  window.open(url, '_blank').focus();
}

function titleCase(s) {
  return s.toLowerCase()
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

function prettyDate(pubDate) {
  const date = new Date(pubDate);

  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const year = date.getFullYear();
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');

  return `${day}/${month}/${year} ${hours}:${minutes}`;
}
