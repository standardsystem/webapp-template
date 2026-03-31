import { authApi } from '@/lib/api'

const providers = [
  { key: 'google', label: 'Google でログイン' },
  { key: 'github', label: 'GitHub でログイン' },
  { key: 'microsoft', label: 'Microsoft でログイン' },
] as const

export function LoginPage() {
  const handleLogin = (provider: string) => {
    window.location.href = authApi.loginURL(provider)
  }

  return (
    <main style={{ maxWidth: 400, margin: '80px auto', textAlign: 'center' }}>
      <h1>Webapp Template</h1>
      <p>ログインしてください</p>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginTop: 24 }}>
        {providers.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => handleLogin(key)}
            style={{
              padding: '12px 24px',
              fontSize: 16,
              cursor: 'pointer',
              border: '1px solid #ccc',
              borderRadius: 8,
              background: '#fff',
            }}
          >
            {label}
          </button>
        ))}
      </div>
    </main>
  )
}
