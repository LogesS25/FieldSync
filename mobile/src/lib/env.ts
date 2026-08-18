import Constants from 'expo-constants';

// Falls back to the local backend dev port (see backend/.env.example) when
// EXPO_PUBLIC_API_URL is not set, so `expo start` works out of the box.
export const API_URL: string =
  process.env.EXPO_PUBLIC_API_URL ??
  (Constants.expoConfig?.extra?.apiUrl as string | undefined) ??
  'http://localhost:8090';
