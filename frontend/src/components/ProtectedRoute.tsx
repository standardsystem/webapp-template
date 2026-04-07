import { type ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";

type Props = {
  children: ReactNode;
  requiredRole?: "admin" | "member";
};

export function ProtectedRoute({ children, requiredRole }: Props) {
  const { user, loading } = useAuth();

  if (loading) {
    return <div className="mt-20 text-center">読み込み中...</div>;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (requiredRole === "admin" && user.role !== "admin") {
    return (
      <div className="mt-20 text-center">
        <h2 className="text-xl font-bold">アクセス権限がありません</h2>
        <p className="mt-2 text-gray-600">この機能には管理者権限が必要です。</p>
      </div>
    );
  }

  return <>{children}</>;
}
