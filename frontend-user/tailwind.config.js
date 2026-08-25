/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        void: "#07090c",
        panel: "#121820",
        rail: "#2a3544",
        fog: "#9aa8b8",
        paper: "#e8edf2",
        phosphor: "#3ee88a",
        sodium: "#f5c14a",
        alarm: "#ff4b63",
        cyan: "#3fd4e8",
        violet: "#8b7cff",
        ash: "#6b7583",
      },
      fontFamily: {
        display: ["Teko", "sans-serif"],
        mono: ["IBM Plex Mono", "ui-monospace", "monospace"],
      },
    },
  },
  plugins: [],
};
