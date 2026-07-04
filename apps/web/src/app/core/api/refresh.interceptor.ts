import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, switchMap, throwError, BehaviorSubject, filter, take } from 'rxjs';
import { AuthService } from '../auth/auth.service';

let isRefreshing = false;
const refreshedToken$ = new BehaviorSubject<string | null>(null);

export const refreshInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status !== 401 || req.url.includes('/api/auth/refresh')) {
        return throwError(() => error);
      }

      if (!isRefreshing) {
        isRefreshing = true;
        refreshedToken$.next(null);

        return auth.refreshToken().pipe(
          switchMap((data) => {
            isRefreshing = false;
            refreshedToken$.next(data.accessToken);

            const retriedReq = req.clone({
              setHeaders: { Authorization: `Bearer ${data.accessToken}` },
            });
            return next(retriedReq);
          }),
          catchError((refreshError) => {
            isRefreshing = false;
            auth.logout();
            return throwError(() => refreshError);
          }),
        );
      }

      return refreshedToken$.pipe(
        filter((token) => token !== null),
        take(1),
        switchMap((token) => {
          const retriedReq = req.clone({
            setHeaders: { Authorization: `Bearer ${token}` },
          });
          return next(retriedReq);
        }),
      );
    }),
  );
};
