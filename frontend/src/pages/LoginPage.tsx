import { authApi } from "@/lib/api";

const providers = [
  { key: "google", label: "Google でログイン" },
  { key: "github", label: "GitHub でログイン" },
  { key: "microsoft", label: "Microsoft でログイン" },
] as const;

export function LoginPage() {
  const handleLogin = (provider: string) => {
    window.location.href = authApi.loginURL(provider);
  };

  return (
    <main className="mx-auto mt-20 max-w-sm text-center">
      <h1 className="mb-2 text-2xl font-bold">Webapp Template</h1>
      <p className="mb-6 text-gray-600">ログインしてください</p>
      <div className="flex flex-col gap-3">
        {providers.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => handleLogin(key)}
            className="cursor-pointer rounded-lg border border-gray-300 bg-white px-6 py-3 text-base hover:bg-gray-50"
          >
            {label}
          </button>
        ))}
      </div>
    </main>
  );
}
