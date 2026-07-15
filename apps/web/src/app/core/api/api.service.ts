import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);
  private baseUrl = environment.apiUrl;

  get<T>(path: string, params?: Record<string, string>) {
    const requestParams = { ...(params ?? {}), _ts: Date.now().toString() };

    return this.http
      .get<ApiResponse<T>>(`${this.baseUrl}${path}`, {
        params: requestParams,
        withCredentials: true,
      })
      .pipe(map((res) => res.data));
  }

  post<T>(path: string, body: unknown) {
    return this.http
      .post<ApiResponse<T>>(`${this.baseUrl}${path}`, body, { withCredentials: true })
      .pipe(map((res) => res.data));
  }

  put<T>(path: string, body: unknown) {
    return this.http
      .put<ApiResponse<T>>(`${this.baseUrl}${path}`, body, { withCredentials: true })
      .pipe(map((res) => res.data));
  }

  delete<T>(path: string) {
    return this.http
      .delete<ApiResponse<T>>(`${this.baseUrl}${path}`, { withCredentials: true })
      .pipe(map((res) => res.data));
  }

  postFormData<T>(path: string, formData: FormData) {
    return this.http
      .post<ApiResponse<T>>(`${this.baseUrl}${path}`, formData, { withCredentials: true })
      .pipe(map((res) => res.data));
  }
}
