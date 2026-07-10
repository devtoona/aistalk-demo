"use client";

import dynamic from "next/dynamic";

const ChatPageContent = dynamic(() => import("./chat/ChatPageContent"), { ssr: false });

export default function Home() {
	return <ChatPageContent />;
}
