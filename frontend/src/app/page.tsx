import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { Landing } from "@/components/landing/landing";
import { AUTH_REFRESH_COOKIE_NAME } from "@/lib/auth/server-constants";

export default async function HomePage() {
  const cookieStore = await cookies();
  const hasRefreshCookie = Boolean(cookieStore.get(AUTH_REFRESH_COOKIE_NAME)?.value);

  // Authenticated users go straight to the app; everyone else sees the landing.
  if (hasRefreshCookie) {
    redirect("/app");
  }

  return <Landing />;
}
