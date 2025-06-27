import { jwtDecode } from 'jwt-decode';

export interface JWTPayload {
  exp: number;
  iat: number;
  sub: string;
}

export function isTokenExpired(token: string): boolean {
  try {
    const decoded = jwtDecode<JWTPayload>(token);
    const currentTime = Math.floor(Date.now() / 1000);
    // Check if token expires in the next 5 minutes (300 seconds)
    return decoded.exp - currentTime < 300;
  } catch (error) {
    console.error('Token decode error:', error);
    return true;
  }
}
