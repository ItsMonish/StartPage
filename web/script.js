window.onload = function() {
  document.getElementById("search-query").addEventListener("keypress", function(ev) {
    if (ev.key == "Enter") searchDDG();
  });
  document.getElementById("category-filter").addEventListener("change", function(ev) {
    activateFilter();
  });
  document.getElementById("read-filter").addEventListener("change", function(ev) {
    activateFilter();
  })
  document.getElementById("list-read").addEventListener("click", function(ev) {
    markListRead();
  })
  document.getElementById("read-filter").selectedIndex = 0;
  renderSources();
  renderRSS();
}

function searchDDG() {
  let query = document.getElementById("search-query").value;
  window.location.href = "https://duckduckgo.com/?q=" + query;
  document.getElementById("search-query").value = "";
}

function markListRead() {
  let filter = document.getElementById("category-filter").value;
  let readFilter = document.getElementById("read-filter").value;

  if (readFilter != "unread") return;

  fetch("/rss/" + filter + "/readAll")
    .then(response => {
      if (!response.ok) {
        throw new Error("Error collection RSS feed with filter " + val);
      } else {
        let feedList = document.getElementById("feed-list");
        let notificationBubble = document.getElementById("rss-bubble");
        notificationBubble.innerText = notificationBubble.innerText - feedList.children.length;
        feedList.innerHTML = "";
        let nothinghere = document.createElement("h2")
        nothinghere.style = "margin-top: 3%; margin-left: 1%";
        if (readFilter == "unread")
          nothinghere.innerText = "It seems you have read it all...";
        else if (readFilter == "favourite")
          nothinghere.innerText = "It seems you haven't favourited anything...";
        else
          nothinghere.innerText = "It seems there is nothing here...";
        feedList.appendChild(nothinghere);
      }
    })
    .catch(error => {
      console.log(error)
    });
}

function activateFilter() {
  let filter = document.getElementById("category-filter").value;
  let val = document.getElementById("read-filter").value;
  let reqUrl = "";
  if (val == "unread") {
    reqUrl = "/rss/" + filter;
  } else if (val == "read") {
    reqUrl = "/rss/" + filter + "/viewed";
  } else if (val == "favourites") {
    //reqUrl = "/rss/" + filter + "/favourites";
    console.log("yet to implement"); return;
  } else {
    console.error("Invalid read filter passed");
  }
  fetch(reqUrl)
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

  if (jsonRssFeed == null || jsonRssFeed.length == 0) {
    let readFilter = document.getElementById("read-filter").value;
    let nothinghere = document.createElement("h2")
    nothinghere.style = "margin-top: 3%; margin-left: 1%";
    if (readFilter == "unread")
      nothinghere.innerText = "It seems you have read it all...";
    else if (readFilter == "favourite")
      nothinghere.innerText = "It seems you haven't favourited anything...";
    else
      nothinghere.innerText = "It seems there is nothing here...";
    feedList.appendChild(nothinghere);
    if (readFilter == "unread") document.getElementById("rss-bubble").innerText = 0;
    return;
  }

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
    newNode.appendChild(header);
    newNode.appendChild(src);
    newNode.appendChild(pubdate);
    if ("readAt" in curObj) {
      let readAt = document.createElement("p");
      readAt.innerHTML = "<strong>Read At:</strong> " + prettyDate(curObj.readAt);
      newNode.appendChild(readAt);
      newNode.setAttribute("onclick", `window.open("${curObj.link}", '_blank').focus();`);
      let favourite = document.createElement("button");
      favourite.classList.add("right-button");
      if (!curObj.isFavourite) {
        favourite.title = "Mark as Favourite";
        favourite.setAttribute("onclick", `event.stopPropagation();markAsFavourite(this, "${curObj.link}")`);
        favourite.innerHTML = `&#x2606`;
      } else {
        favourite.title = "Unmark as Favourite";
        favourite.setAttribute("onclick", `event.stopPropagation();unmarkAsFavourite(this, "${curObj.link}")`);
        favourite.innerHTML = `&#x2605`;
      }
      newNode.appendChild(favourite);
      markIcon.innerHTML = `&#x2717`;
      markIcon.title = "Mark as Unread";
      markIcon.style = "right: 120px;"
      markIcon.classList.add("right-button");
      markIcon.setAttribute("onclick", "event.stopPropagation();console.log('do it')");
      newNode.appendChild(markIcon);
    } else {
      markIcon.innerHTML = `&#x1F441`;
      markIcon.setAttribute("onclick", `event.stopPropagation();markAsRead(this, ${curObj.id}, true);`);
      markIcon.setAttribute("title", "Mark as Read");
      markIcon.classList.add("right-button")
      newNode.appendChild(markIcon);
      newNode.setAttribute("onclick", `newTab(this,"${curObj.id}" ,"${curObj.link}");`);
    }
    feedList.appendChild(newNode);
  }
}

function markAsFavourite(caller, link) {
  fetch("/rss/item/favourite", {
    method: "POST",
    body: link
  })
    .then(response => {
      if (!response.ok) {
        throw new Error("Error marking item as favourite");
      }
      caller.innerHTML = `&#x2605`;
      caller.title = "Unmark as Favourite";
      caller.setAttribute("onclick", `event.stopPropagation(); unmarkAsFavourite(this, "${link}");`)
    })
    .catch(error => {
      console.log(error)
    });
}

function unmarkAsFavourite(caller, link) {
  fetch("/rss/item/unfavourite", {
    method: "POST",
    body: link
  })
    .then(response => {
      if (!response.ok) {
        throw new Error("Error unmarking item as favourite");
      }
      caller.innerHTML = `&#x2606`;
      caller.title = "Mark as Favourite";
      caller.setAttribute("onclick", `event.stopPropagation(); markAsFavourite(this, "${link}");`)
    })
    .catch(error => {
      console.log(error)
    });
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
  pubDate = pubDate.split(" ").slice(0, 3).join(" ");
  const date = new Date(pubDate);

  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const year = date.getFullYear();
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');

  return `${day}/${month}/${year} ${hours}:${minutes}`;
}
