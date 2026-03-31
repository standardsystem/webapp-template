import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { UserList } from './UserList'
import type { User } from '@/lib/api'

const mockUsers: User[] = [
  { id: '1', name: '加藤一由樹', email: 'kazuyuki@example.com', createdAt: '', updatedAt: '' },
  { id: '2', name: '山田太郎', email: 'yamada@example.com', createdAt: '', updatedAt: '' },
]

describe('UserList', () => {
  it('ローディング中はステータスを表示する', () => {
    render(<UserList users={[]} loading={true} error={null} />)
    expect(screen.getByRole('status')).toHaveTextContent('読み込み中...')
  })

  it('エラー時はエラーメッセージを表示する', () => {
    render(<UserList users={[]} loading={false} error="通信エラー" />)
    expect(screen.getByRole('alert')).toHaveTextContent('通信エラー')
  })

  it('ユーザーが0件のときはメッセージを表示する', () => {
    render(<UserList users={[]} loading={false} error={null} />)
    expect(screen.getByText('ユーザーがいません')).toBeInTheDocument()
  })

  it('ユーザー一覧を正しく表示する', () => {
    render(<UserList users={mockUsers} loading={false} error={null} />)
    expect(screen.getByText('加藤一由樹')).toBeInTheDocument()
    expect(screen.getByText('山田太郎')).toBeInTheDocument()
    expect(screen.getAllByRole('listitem')).toHaveLength(2)
  })
})
