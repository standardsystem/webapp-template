import { useAuth } from "@/contexts/AuthContext";
import { UserList } from "@/components/UserList";
import { useUsers } from "@/hooks/useUsers";

export function DashboardPage() {
  const { user, logout } = useAuth();
  const { users, loading, error, refetch } = useUsers();

  return (
    <main className="mx-auto max-w-3xl p-6">
      <header className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Webapp Template</h1>
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-600">
            {user?.name}（{user?.role}）
          </span>
          <button
            onClick={logout}
            className="rounded bg-gray-200 px-3 py-1 text-sm hover:bg-gray-300"
          >
            ログアウト
          </button>
        </div>
      </header>
      <button
        onClick={refetch}
        disabled={loading}
        className="mb-4 rounded bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
      >
        更新
      </button>
      <UserList users={users} loading={loading} error={error} />
    </main>
  );
}
