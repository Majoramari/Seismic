import { Component, inject, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { Navbar } from './shared/components/navbar/navbar';
import { Sidebar } from './shared/components/sidebar/sidebar';
import { ToastContainer } from './shared/components/toast/toast';
import { AuthService } from './core/auth/auth.service';
import { UserService } from './core/auth/user.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, Navbar, Sidebar, ToastContainer],
  templateUrl: './app.html',
})
export class App implements OnInit {
  auth = inject(AuthService);
  private userService = inject(UserService);

  ngOnInit() {
    if (this.auth.isLoggedIn()) {
      this.userService.load();
    }
  }
}
