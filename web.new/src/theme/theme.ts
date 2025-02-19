export enum AppTheme {
  DARK = 0,
  LIGHT = 1,
}

export const DarkTheme = {
  background: '#111417',
  background2: '#191c20',
  background3: '#25282c',

  text: '#f4f4f5',
  textAlt: '#f4f4f5',

  accent: '#fe281f',
  accentDarker: '#b33a2d',

  white: '#f4f4f5',
  whiteDarker: '#dddddd',
  blurple: '#5865f2',
  blurpleDarker: '#4450d6',
  darkGray: '#1e1e1e',
  red: '#ed4245',
  orange: '#f57c00',
  yellow: '#fbc02d',
  green: '#43a047',
  lime: '#57f287',
  cyan: '#03a9f4',
  pink: '#eb459e',
};

export const LightTheme: Theme = {
  ...DarkTheme,

  background: '#fffffe',
  background2: '#e2e2e2',
  background3: '#d0d0d0',

  text: '#212121',
  textAlt: '#f4f4f5',

  accentDarker: '#fd756f',
};

export const DefaultTheme = DarkTheme;
export type Theme = typeof DefaultTheme;

export const getSystemTheme = () => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? AppTheme.DARK : AppTheme.LIGHT;
};
