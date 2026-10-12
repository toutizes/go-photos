// Existing Flutter clients also receive a cleanup worker at the old script URL.
// Here, remove registrations if the new page has already loaded successfully.
export async function retireFlutter() {
  if (!("serviceWorker" in navigator)) return;
  const scope = new URL(import.meta.env.BASE_URL, location.origin).href;
  for (const registration of await navigator.serviceWorker.getRegistrations()) {
    const worker =
      registration.active ?? registration.waiting ?? registration.installing;
    if (
      registration.scope === scope &&
      worker?.scriptURL.includes("/flutter_service_worker.js")
    ) {
      await registration.unregister();
      if ("caches" in window)
        await Promise.all(
          [
            "flutter-app-cache",
            "flutter-temp-cache",
            "flutter-app-manifest",
          ].map((name) => caches.delete(name)),
        );
    }
  }
}
