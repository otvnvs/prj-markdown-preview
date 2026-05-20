// -----------------------------------------------------------------------------
// PWA
// -----------------------------------------------------------------------------
const cacheName = "app-" + "ae71d09e51bd3e98097dea94b2f4c2ed3fc396d8";
const resourcesToCache = ["https://raw.githubusercontent.com/maxence-charriere/go-app/master/docs/web/icon.png","/prj-markdown-preview/web/style.css","/prj-markdown-preview/web/app.wasm","/prj-markdown-preview/wasm_exec.js","/prj-markdown-preview/manifest.webmanifest","/prj-markdown-preview/app.js","/prj-markdown-preview/app.css","/prj-markdown-preview"];

self.addEventListener("install", async (event) => {
  try {
    console.log("installing app worker ae71d09e51bd3e98097dea94b2f4c2ed3fc396d8");
    await installWorker();
    await self.skipWaiting();
  } catch (error) {
    console.error("error during installation:", error);
  }
});

async function installWorker() {
  const cache = await caches.open(cacheName);
  await cache.addAll(resourcesToCache);
}

self.addEventListener("activate", async (event) => {
  try {
    await deletePreviousCaches(); // Await cache cleanup
    await self.clients.claim(); // Ensure the service worker takes control of the clients
    console.log("app worker ae71d09e51bd3e98097dea94b2f4c2ed3fc396d8 is activated");
  } catch (error) {
    console.error("error during activation:", error);
  }
});

async function deletePreviousCaches() {
  const keys = await caches.keys();
  await Promise.all(
    keys.map(async (key) => {
      if (key !== cacheName) {
        try {
          console.log("deleting", key, "cache");
          await caches.delete(key);
        } catch (err) {
          console.error("deleting", key, "cache failed:", err);
        }
      }
    })
  );
}

self.addEventListener("fetch", (event) => {
  event.respondWith(fetchWithCache(event.request));
});

async function fetchWithCache(request) {
  const cachedResponse = await caches.match(request);
  if (cachedResponse) {
    return cachedResponse;
  }
  return await fetch(request);
}

// -----------------------------------------------------------------------------
// Push Notifications
// -----------------------------------------------------------------------------
self.addEventListener("push", (event) => {
  event.waitUntil((async () => {
    let notification;

    try {
      notification = event.data ? event.data.json() : null;
    } catch {
      notification = null;
    }

    if (!notification) {
      return;
    }

    await showNotification(self.registration, notification);
  })());
});

self.addEventListener("message", (event) => {
  const msg = event.data;
  if (!msg || msg.type !== "goapp:notify") {
    return;
  }

  event.waitUntil(
    showNotification(self.registration, msg.options)
  );
});

async function showNotification(registration, notification) {
  const title = notification.title || "Notification";

  let actions = [];
  for (let i in notification.actions) {
    const action = notification.actions[i];
    actions.push({
      action: action.action,
      path: action.path,
    });
    delete action.path;
  }

  await registration.showNotification(title, {
    body: notification.body,
    icon: notification.icon,
    badge: notification.badge,
    actions: notification.actions,
    data: {
      goapp: {
        path: notification.path,
        actions: actions
      }
    }
  });
}

self.addEventListener("notificationclick", (event) => {
  event.notification.close();

  const notification = event.notification;
  let path = notification.data.goapp.path;

  for (let i in notification.data.goapp.actions) {
    const action = notification.data.goapp.actions[i];
    if (action.action === event.action) {
      path = action.path;
      break;
    }
  }

  event.waitUntil(
    clients
      .matchAll({
        type: "window",
      })
      .then((clientList) => {
        for (var i = 0; i < clientList.length; i++) {
          let client = clientList[i];
          if ("focus" in client) {
            client.focus();
            client.postMessage({
              goapp: {
                type: "notification",
                path: path,
              },
            });
            return;
          }
        }

        if (clients.openWindow) {
          return clients.openWindow(path);
        }
      })
  );
});