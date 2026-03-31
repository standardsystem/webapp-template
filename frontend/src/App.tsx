import { AuthProvider, useAuth } from '@/contexts/AuthContext'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { UserList } from '@/components/UserList'
import { useUsers } from '@/hooks/useUsers'

function Dashboard() {
  const { user, logout } = useAuth()
  const { users, loading, error, refetch } = useUsers()

  return (
    <main>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>Webapp Template</h1>
        <div>
          <span>{user?.name}（{user?.role}）</span>
          <button onClick={logout} style={{ marginLeft: 12 }}>
            ログアウト
          </button>
        </div>
      </header>
      <button onClick={refetch} disabled={loading}>
        更新
      </button>
      <UserList users={users} loading={loading} error={error} />
    </main>
  )
}

function App() {
  return (
    <AuthProvider>
      <ProtectedRoute>
        <Dashboard />
      </ProtectedRoute>
    </AuthProvider>
  )
}

export default App
