import { createRoot } from "react-dom/client";
import { App } from "./App";
import { ErrorBoundary } from "./ErrorBoundary";
import { retireFlutter } from "./retireFlutter";
import "@fontsource-variable/manrope";
import "@fontsource-variable/dm-sans";
import "./style.css";
createRoot(document.getElementById("root")!).render(
  <ErrorBoundary>
    <App />
  </ErrorBoundary>,
);
void retireFlutter().catch(() => {
  /* Browsing remains usable when storage is unavailable. */
});
