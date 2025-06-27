import { NextRequest, NextResponse } from "next/server";

export async function middleware(request: NextRequest) {
  // Pass through to the next middleware or route handler
  return NextResponse.next();
}

export const config = {
  /*
   * Match all request paths except for the ones starting with:
   * - _next/static (static files)
   * - _next/image (image optimization files)
   * - favicon.ico (favicon file)
   */
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};