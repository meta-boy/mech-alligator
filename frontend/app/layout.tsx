import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";
import { ThemeProvider } from "../components/ui/ThemeProvider";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
  weight: "100 900",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export const metadata: Metadata = {
  title: "Mech Alligator",
  description: "A one-stop shop for all your mechanical keyboard needs",
  openGraph: {
    title: "Mech Alligator",
    description: "Discover premium keyboards and keycaps from top resellers",
    url: "https://agg.regator.site/",
    siteName: "Mech Alligator",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Mech Alligator - Premium Mechanical Keyboards",
      },
    ],
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Mech Alligator",
    description: "Discover premium keyboards and keycaps from top resellers",
    images: ["/og-image.png"],
  },
  metadataBase: new URL("https://agg.regator.site/"),
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider>
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}
