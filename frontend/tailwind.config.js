/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'siem-dark': '#0a0e17',
        'siem-card': '#111827',
        'siem-border': '#1f2937',
        'alert-red': '#dc2626',
      }
    },
  },
  plugins: [],
}
