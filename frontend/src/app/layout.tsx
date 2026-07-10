import type { Metadata } from "next";
import { M_PLUS_Rounded_1c, Nunito } from "next/font/google";
import "./globals.css";
import { TTSLoadingProvider } from "@/contexts/TTSLoadingContext";

const mPlusRounded = M_PLUS_Rounded_1c({
  subsets: ["latin"],
  weight: ["400", "500", "700"],
  variable: "--font-rounded",
  display: "swap",
});

const nunito = Nunito({
  subsets: ["latin"],
  weight: ["600", "700"],
  variable: "--font-nunito",
  display: "swap",
});

export const metadata: Metadata = {
  title: process.env.NEXT_PUBLIC_APP_NAME ?? "AISTalk",
  description: "Voice chat with AI and VRM avatar",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja" className={`${mPlusRounded.variable} ${nunito.variable}`}>
      <body className="antialiased font-rounded">
        <TTSLoadingProvider>{children}</TTSLoadingProvider>
      </body>
    </html>
  );
}
