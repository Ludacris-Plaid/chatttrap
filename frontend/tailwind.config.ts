import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        cyber: {
          bg: "#0a0a0a",
          card: "#111111",
          border: "#1a1a2e",
          red: "#dc2626",
          "red-dim": "#991b1b",
          green: "#22c55e",
          text: "#e5e5e5",
          muted: "#737373",
        },
      },
      fontFamily: {
        mono: ["JetBrains Mono", "Fira Code", "monospace"],
      },
      animation: {
        glow: "glow 2s ease-in-out infinite alternate",
        pulse: "pulse 2s ease-in-out infinite",
      },
      keyframes: {
        glow: {
          "0%": { boxShadow: "0 0 5px rgba(220, 38, 38, 0.5)" },
          "100%": { boxShadow: "0 0 20px rgba(220, 38, 38, 0.8)" },
        },
      },
    },
  },
  plugins: [],
};

export default config;
