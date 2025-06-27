import { jwtDecode } from "jwt-decode";
import { isTokenExpired, JWTPayload } from "./jwt";

export class AuthManager {
  private static instance: AuthManager;
  private token: string | null = null;

  private constructor() {}

  static getInstance(): AuthManager {
    if (!AuthManager.instance) {
      AuthManager.instance = new AuthManager();
    }
    return AuthManager.instance;
  }

  async getToken(): Promise<string | null> {
    // Try to get token from memory first
    if (this.token && !this.isTokenExpired(this.token)) {
      return this.token;
    }

    // Try to get token from cookie via API route
    try {
      const response = await fetch('/api/auth/token');
      if (response.ok) {
        const data = await response.json();
        this.token = data.token;
        return this.token;
      }
    } catch (error) {
      console.error('Error getting token:', error);
    }

    return null;
  }

  private isTokenExpired(token: string): boolean {
    try {
      const decoded = jwtDecode<JWTPayload>(token);
      const currentTime = Math.floor(Date.now() / 1000);
      return decoded.exp - currentTime < 300; // 5 minutes buffer
    } catch (error) {
      return true;
    }
  }

  async makeAuthenticatedRequest(url: string, options: RequestInit = {}): Promise<Response> {
    const token = await this.getToken();
    
    if (!token) {
      throw new Error('No valid token available');
    }

    const headers = new Headers(options.headers);
    headers.set('Authorization', `Bearer ${token}`);

    return fetch(url, {
      ...options,
      headers,
    });
  }
}