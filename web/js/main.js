const mousedown = event => {
  event.target.classList.add("clicked");
}

const showOverlay = () => {
  let overlay = document.querySelector('div#overlay');
  overlay.style.display = 'block';
}

const hideOverlay = () => {
  let overlay = document.querySelector('div#overlay');
  overlay.style.display = 'none';
}

const changeVPN = async profile => {
  document.querySelectorAll('div.active').forEach(async (e) => {
    if(profile != e.dataset['profile']) {
      await fetch(`/vpn/disconnect?profile=${e.dataset["profile"]}`);
    }
  });
  await fetch(`/vpn/connect?profile=${profile}`);
}

const mouseup = event => {
  let target = event.target;
  target.classList.remove("clicked");

  let flag = document.querySelector('div#flag');
  flag.innerHTML = target.dataset['flag'];

  showOverlay();
  if(target.classList.contains('active')) {
    fetch(`/vpn/disconnect?profile=${target.dataset["profile"]}`).then(() => {
      hideOverlay();
    });
  }
  else {
    changeVPN(target.dataset['profile']);
  }
}

const vpnConnect = event => {
  let elem = document.querySelector(`div[data-profile="${event.data}"]`);
  elem.classList.add("active");
  hideOverlay();
}

const vpnDisconnect = event => {
  let elem = document.querySelector(`div[data-profile="${event.data}"]`);
  elem.classList.remove("active");
}

let vpnList = fetch('/vpn/list').then((response) => {
    return response.json();
});

window.addEventListener('load', function () {
  vpnList.then((vpnProfiles) => {
    let vpnListElement = document.querySelector("div#vpn_list");

    let profiles = vpnProfiles.profiles;

    Object.keys(profiles).forEach((k) => {
      let button = document.createElement("div");

      button.setAttribute("class", "quick_connect");
      button.setAttribute("data-profile", k);
      button.setAttribute("data-flag", profiles[k].flag);

      if(profiles[k].running) {
        button.classList.add("active");
      }

      button.innerHTML = profiles[k].flag + "<br/>" + profiles[k].city;
      button.addEventListener("mousedown", mousedown);
      button.addEventListener("mouseup", mouseup);
      vpn_list.appendChild(button);
    });
  });

  let vpnStream = new EventSource('/vpn/stream');
  vpnStream.addEventListener("connect", vpnConnect);
  vpnStream.addEventListener("disconnect", vpnDisconnect);
});

