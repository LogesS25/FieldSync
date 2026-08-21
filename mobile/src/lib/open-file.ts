import { Platform } from 'react-native';
import * as FileSystem from 'expo-file-system/legacy';
import * as Sharing from 'expo-sharing';

// The daily-report file endpoint requires an Authorization header, so it
// can't be opened as a plain URL — download it (native) or fetch it as a
// blob (web) with the current access token, then hand it to the platform's
// viewer.
export async function openAuthenticatedFile(url: string, accessToken: string, filename: string): Promise<void> {
  if (Platform.OS === 'web') {
    const response = await fetch(url, { headers: { Authorization: `Bearer ${accessToken}` } });
    if (!response.ok) throw new Error('Could not download file');
    const blob = await response.blob();
    const objectUrl = URL.createObjectURL(blob);
    window.open(objectUrl, '_blank');
    return;
  }

  const localUri = FileSystem.cacheDirectory + filename;
  const { uri } = await FileSystem.downloadAsync(url, localUri, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (await Sharing.isAvailableAsync()) {
    await Sharing.shareAsync(uri);
  }
}
