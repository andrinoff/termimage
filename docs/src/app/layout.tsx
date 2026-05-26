import type { Metadata } from "next";
import "./globals.css";
import "highlight.js/styles/github-dark.css";

export const metadata: Metadata = {
	title: "termimage",
	description:
		"Sandboxed terminal image rendering for Go. Kitty, Sixel, and half-block. Fast CGo decode via stb_image.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en">
			<body>
				<header className="site-header">
					<a href="/" className="brand">
						termimage
					</a>
					<nav>
						<a href="https://github.com/floatpane/termimage">GitHub</a>
					</nav>
				</header>
				<main>{children}</main>
			</body>
		</html>
	);
}
