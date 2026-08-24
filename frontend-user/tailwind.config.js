/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts}'],
  theme: {
    extend: {
      colors: {
        bg: '#12180F',
        panel: '#1B2416',
        ink: '#E8E1D1',
        muted: '#9AA48A',
        pine: '#3F6B4A',
        moss: '#7C9A5A',
        clay: '#C47A3A',
        danger: '#D4533A',
        water: '#4E8FA8',
        line: '#2A3324',
      },
      fontFamily: {
        display: ['Fraunces', 'Iowan Old Style', 'serif'],
        body: ['"Noto Serif SC"', 'Iowan Old Style', 'serif'],
      },
    },
  },
  plugins: [],
}
