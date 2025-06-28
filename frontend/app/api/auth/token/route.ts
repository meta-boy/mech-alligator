import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { API_BASE_URL, TOKEN_COOKIE_NAME } from "@/utils/constants";
import { isTokenExpired } from "@/utils/jwt";

async function loginAndGetToken(): Promise<string | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        username: process.env.API_USERNAME,
        password: process.env.API_PASSWORD,
      }),
    });

    if (!response.ok) {
      console.error("Login failed:", response.status, response.statusText);
      return null;
    }

    const data = await response.json();
    return data.token;
  } catch (error) {
    console.error("Login error:", error);
    return null;
  }
}

export async function GET() {
  const cookieStore = cookies();
  let token = cookieStore.get(TOKEN_COOKIE_NAME)?.value;
  let needsNewToken = false;

  if (!token || isTokenExpired(token)) {
    const newToken = await loginAndGetToken();
    if (newToken) {
      token = newToken;
      needsNewToken = true;
    } else {
      return NextResponse.json({ message: "Authentication failed" }, { status: 401 });
    }
  }

  const response = NextResponse.json({ token });

  if (needsNewToken && token) {
    response.cookies.set(TOKEN_COOKIE_NAME, token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "strict",
      maxAge: 24 * 60 * 60, // 24 hours
    });
  }

  return response;
}
