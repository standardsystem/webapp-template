import { type ReactNode } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { LoginPage } from '@/pages/LoginPage'

type Props = {
  children: ReactNode
  requiredRole?: 'admin' | 'member'
}

export function ProtectedRoute({ children, requiredRole }: Props) {
  const { user, loading } = useAuth()

  if (loading) {
    return <div style={{ textAlign: 'center', marginTop: 80 }}>読み込み中...</div>
  }

  if (!user) {
    return <LoginPage />
  }

  if (requiredRole === 'admin' && user.role !== 'admin') {
    return (
      <div style={{ textAlign: 'center', marginTop: 80 }}>
        <h2>アクセス権限がありません</h2>
        <p>この機能には管理者権限が必要です。</p>
      </div>
    )
  }

  return <>{children}</>
}
