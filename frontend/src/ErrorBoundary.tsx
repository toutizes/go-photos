import { Component, type ErrorInfo, type ReactNode } from "react";
export class ErrorBoundary extends Component<
  { children: ReactNode },
  { failed: boolean }
> {
  state = { failed: false };
  static getDerivedStateFromError() {
    return { failed: true };
  }
  componentDidCatch(error: Error, _info: ErrorInfo) {
    console.error("Erreur de l’interface Toutizes", error);
  }
  render() {
    return this.state.failed ? (
      <main className="empty">
        <h1>Un problème est survenu</h1>
        <p>Rechargez la page pour retrouver vos photos.</p>
        <button onClick={() => location.reload()}>Recharger</button>
      </main>
    ) : (
      this.props.children
    );
  }
}
