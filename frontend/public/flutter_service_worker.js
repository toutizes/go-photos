// Retire the previous Flutter offline shell when an existing installation updates.
// This worker never intercepts requests or caches authenticated photos.
self.addEventListener("install", () => self.skipWaiting());
self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      await Promise.all(
        ["flutter-app-cache", "flutter-temp-cache", "flutter-app-manifest"].map(
          (name) => caches.delete(name),
        ),
      );
      await self.registration.unregister();
      const windows = await self.clients.matchAll({
        type: "window",
        includeUncontrolled: true,
      });
      await Promise.all(
        windows
          .filter((client) => client.url.startsWith(self.registration.scope))
          .map((client) => client.navigate(client.url)),
      );
    })(),
  );
});
