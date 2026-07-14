export type TutorialSlide = {
	id: number;
	title: string;
	description: string;
	note: string;
};

type TutorialDataFile = {
	tutorial: TutorialSlide[];
};

export const TUTORIAL_DISMISS_LS_KEY = "aistalk_demo_tutorial_dismissed_v1";

export function isTutorialDismissed(): boolean {
	if (typeof window === "undefined") return true;
	try {
		return localStorage.getItem(TUTORIAL_DISMISS_LS_KEY) === "1";
	} catch {
		return false;
	}
}

export function setTutorialDismissed(dismissed: boolean) {
	if (typeof window === "undefined") return;
	try {
		if (dismissed) {
			localStorage.setItem(TUTORIAL_DISMISS_LS_KEY, "1");
		} else {
			localStorage.removeItem(TUTORIAL_DISMISS_LS_KEY);
		}
	} catch {
		// ignore quota errors
	}
}

export async function fetchTutorialSlides(): Promise<TutorialSlide[]> {
	const res = await fetch("/tutorial/data.json", { cache: "no-store" });
	if (!res.ok) {
		throw new Error(`tutorial data fetch failed: ${res.status}`);
	}
	const data = (await res.json()) as TutorialDataFile;
	if (!Array.isArray(data.tutorial)) {
		throw new Error("tutorial data is invalid");
	}
	return data.tutorial;
}

export function tutorialImageSrc(slideId: number): string {
	return `/tutorial/images/${slideId}.png`;
}
