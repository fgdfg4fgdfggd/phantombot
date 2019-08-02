/** @format */

import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { HttpClient } from '@angular/common/http';
import {
  User,
  ListReponse,
  Guild,
  PermLvlResponse,
  Member,
  Report,
} from './api.models';
import { environment } from 'src/environments/environment';
import { ToastService } from '../components/toast/toast.service';

/** @format */

@Injectable({
  providedIn: 'root',
})
export class APIService {
  private rootURL = '';
  private defopts = {
    withCredentials: true,
  };

  private errorChatcher = (err) => {
    console.error(err);
    this.toasts.push(err.message, 'Request Error', 'error', 10000);
    return of(null);
  };

  constructor(private http: HttpClient, private toasts: ToastService) {
    this.rootURL = environment.production ? '' : 'http://localhost:8080';
  }

  public logout(): Observable<any> {
    return this.http
      .post<any>(this.rootURL + '/api/logout', this.defopts)
      .pipe(catchError(this.errorChatcher));
  }

  public getSelfUser(): Observable<User> {
    return this.http.get<User>(this.rootURL + '/api/me', this.defopts).pipe(
      catchError((err) => {
        if (err.status !== 401) {
          return this.errorChatcher(err);
        }
      })
    );
  }

  public getGuilds(): Observable<Guild[]> {
    return this.http
      .get<ListReponse<Guild>>(this.rootURL + '/api/guilds', this.defopts)
      .pipe(
        map((lr) => {
          return lr.data;
        }),
        catchError(this.errorChatcher)
      );
  }

  public getGuild(id: string): Observable<Guild> {
    return this.http
      .get<Guild>(this.rootURL + '/api/guilds/' + id, this.defopts)
      .pipe(catchError(this.errorChatcher));
  }

  public getGuildMember(guildID: string, memberID: string): Observable<Member> {
    return this.http
      .get<Member>(
        this.rootURL + '/api/guilds/' + guildID + '/' + memberID,
        this.defopts
      )
      .pipe(catchError(this.errorChatcher));
  }

  public getPermissionLvl(guildID: string, userID: string): Observable<number> {
    return this.http
      .get<PermLvlResponse>(
        this.rootURL + '/api/permlvl/' + guildID + '/' + userID,
        this.defopts
      )
      .pipe(
        map((r) => {
          return r.lvl;
        }),
        catchError(this.errorChatcher)
      );
  }

  public getReports(guildID: string, memberID: string): Observable<Report[]> {
    return this.http
      .get<ListReponse<Report>>(
        this.rootURL + '/api/reports/' + guildID + '/' + memberID,
        this.defopts
      )
      .pipe(
        map((lr) => lr.data),
        catchError(this.errorChatcher)
      );
  }

  public getReport(reportID: string): Observable<Report> {
    return this.http
      .get<Report>(this.rootURL + '/api/reports/' + reportID, this.defopts)
      .pipe(catchError(this.errorChatcher));
  }
}