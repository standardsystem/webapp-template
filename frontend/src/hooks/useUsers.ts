import { useCallback, useEffect, useState } from "react";
import { type User, userApi } from "@/lib/api";

type State = {
  users: User[];
  loading: boolean;
  error: string | null;
};

export function useUsers() {
  const [state, setState] = useState<State>({
    users: [],
    loading: false,
    error: null,
  });

  const fetchUsers = useCallback(async () => {
    setState((s) => ({ ...s, loading: true, error: null }));
    try {
      const users = await userApi.list();
      setState({ users, loading: false, error: null });
    } catch (err) {
      setState((s) => ({
        ...s,
        loading: false,
        error: err instanceof Error ? err.message : "Failed to fetch users",
      }));
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  return { ...state, refetch: fetchUsers };
}
