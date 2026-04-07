import type { User } from "@/lib/api";

type Props = {
  users: User[];
  loading: boolean;
  error: string | null;
};

export function UserList({ users, loading, error }: Props) {
  if (loading) {
    return <p role="status" className="text-gray-500">読み込み中...</p>;
  }

  if (error) {
    return (
      <p role="alert" className="text-red-600">
        {error}
      </p>
    );
  }

  if (users.length === 0) {
    return <p className="text-gray-500">ユーザーがいません</p>;
  }

  return (
    <ul className="divide-y divide-gray-200">
      {users.map((user) => (
        <li key={user.id} className="py-3">
          <span className="font-medium">{user.name}</span>
          <span className="ml-2 text-gray-500">{user.email}</span>
        </li>
      ))}
    </ul>
  );
}
