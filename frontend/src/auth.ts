import { initializeApp } from 'firebase/app';
import { getAuth, GoogleAuthProvider, onAuthStateChanged, signInWithPopup, signOut } from 'firebase/auth';
// Public Firebase web configuration, shared with the existing Flutter application.
const app = initializeApp({
  apiKey: 'AIzaSyDYv45ZFi0WMZvSv9WbNc8RAuknm5SrqG8',
  appId: '1:347191427262:web:a229801596c62878a39eef',
  messagingSenderId: '347191427262', projectId: 'floutizes',
  authDomain: 'floutizes.firebaseapp.com',
});
const auth = getAuth(app);
auth.languageCode = 'fr';
export type Account = { email: string | null; displayName: string | null };
export const watchAccount = (fn: (account: Account | null) => void) => onAuthStateChanged(auth, fn);
export const login = () => signInWithPopup(auth, new GoogleAuthProvider());
export const logout = () => signOut(auth);
export const token = (force = false) => auth.currentUser?.getIdToken(force) ?? Promise.resolve(null);
export const isAdmin = (account: Account | null) => account?.email === 'matthieu.devin@gmail.com';
