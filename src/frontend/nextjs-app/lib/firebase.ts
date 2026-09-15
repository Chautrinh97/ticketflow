// Real Firebase Web SDK sign-in, or a mock path when no Firebase project is
// configured (empty NEXT_PUBLIC_FIREBASE_API_KEY) — pairs with
// identity-service's AUTH_FIREBASE_MODE=mock default so the whole stack
// runs end-to-end via `docker compose up` without a provisioned Firebase
// project. See src/services/identity-service/internal/firebase/mock.go for
// the exact token shape identity-service expects in mock mode.
import { initializeApp, type FirebaseApp } from 'firebase/app'
import {
  getAuth,
  GoogleAuthProvider,
  signInWithEmailAndPassword,
  signInWithPopup,
  type Auth,
} from 'firebase/auth'

const firebaseConfig = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID,
}

export const FIREBASE_MOCK_MODE = !firebaseConfig.apiKey

let app: FirebaseApp | null = null
let auth: Auth | null = null

function getFirebaseAuth(): Auth {
  if (!app) app = initializeApp(firebaseConfig)
  if (!auth) auth = getAuth(app)
  return auth
}

export async function signInWithEmailPassword(email: string, password: string): Promise<string> {
  const cred = await signInWithEmailAndPassword(getFirebaseAuth(), email, password)
  return cred.user.getIdToken()
}

export async function signInWithGoogle(): Promise<string> {
  const cred = await signInWithPopup(getFirebaseAuth(), new GoogleAuthProvider())
  return cred.user.getIdToken()
}

// Builds the base64(JSON) token identity-service's MockVerifier decodes —
// only used when FIREBASE_MOCK_MODE is true.
export function buildMockIdToken(email: string, fullName: string): string {
  const uid = `mock-${email}`
  const payload = JSON.stringify({ uid, email, name: fullName })
  return typeof window !== 'undefined' ? window.btoa(payload) : Buffer.from(payload).toString('base64')
}
