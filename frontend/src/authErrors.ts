export function signInErrorMessage(error: unknown): string {
  const rawCode =
    error && typeof error === "object" && "code" in error
      ? error.code
      : undefined;
  const code =
    typeof rawCode === "string" && /^auth\/[a-z-]+$/.test(rawCode)
      ? rawCode
      : "";
  const messages: Record<string, string> = {
    "auth/popup-blocked":
      "Le navigateur a bloqué la fenêtre Google. Autorisez les fenêtres pour ce site puis réessayez.",
    "auth/popup-closed-by-user":
      "La fenêtre Google s’est fermée avant la fin de la connexion. Réessayez.",
    "auth/cancelled-popup-request":
      "Une autre tentative de connexion a remplacé celle-ci. Réessayez.",
    "auth/unauthorized-domain":
      "Ce domaine n’est pas autorisé dans la configuration Firebase.",
    "auth/network-request-failed":
      "Connexion à Google impossible. Vérifiez la connexion et les éventuels bloqueurs du navigateur.",
    "auth/web-storage-unsupported":
      "Le stockage du navigateur est indisponible pour la connexion. Vérifiez les paramètres de confidentialité.",
    "auth/operation-not-allowed":
      "La connexion Google n’est pas activée dans la configuration Firebase.",
  };
  const message = messages[code] ?? "La connexion Google a échoué. Réessayez.";
  return code ? `${message} (${code})` : message;
}
