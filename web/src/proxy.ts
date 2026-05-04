import { NextRequest, NextResponse } from "next/server";

export async function proxy(request: NextRequest) {
  //   const meRes = await fetch(`${backendUrl}/api/v1/auth/me`, {
  //     method: "GET",
  //     headers: {
  //       cookie: request.headers.get("cookie") ?? "",
  //     },
  //     cache: "no-store",
  //   });

  const cookie = request.headers.get("cookie");
  //   return new NextResponse(
  //     JSON.stringify(
  //       {
  //         cookie: request.headers.get("cookie"),
  //         accessToken: request.headers.get("access_token"),
  //       },
  //       null,
  //       2,
  //     ),
  //     {
  //       status: 200,
  //       headers: {
  //         "content-type": "application/json",
  //       },
  //     },
  //   );

  if (!cookie?.includes("access_token")) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("next", request.nextUrl.pathname);
    return NextResponse.redirect(loginUrl);
  }

  //   if (!meRes.ok) {
  //     return new NextResponse(
  //       JSON.stringify(
  //         {
  //           backendUrl,
  //           meStatus: meRes.status,
  //           hasCookie: !!request.headers.get("cookie"),
  //           cookie: request.headers.get("cookie"),
  //           meText: await meRes.text(),
  //         },
  //         null,
  //         2,
  //       ),
  //       {
  //         status: 200,
  //         headers: {
  //           "content-type": "application/json",
  //         },
  //       },
  //     );
  //   }

  //   const body = await meRes.json();
  //   const user = body.data;

  //   if (request.nextUrl.pathname.startsWith("/admin") && user.role !== "admin") {
  //     return NextResponse.redirect(new URL("/dashboard", request.url));
  //   }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/admin/:path*"],
};
