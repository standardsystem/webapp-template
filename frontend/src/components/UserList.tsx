import type { User } from '@/lib/api'

type Props = {
  users: User[]
  loading: boolean
  error: string | null
}

export function UserList({ users, loading, error }: Props) {
  if (loading) {
    return <p role="status">読み込み中...</p>
  }

  if (error) {
    return <p role="alert" style={{ color: 'red' }}>{error}</p>
  }

  if (users.length === 0) {
    return <p>ユーザーがいません</p>
  }

  return (
    <ul>
      {users.map(user => (
        <li key={user.id}>
          <strong>{user.name}</strong> — {user.email}
        </li>
      ))}
    </ul>
  )
}
