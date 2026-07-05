import { Component, inject } from '@angular/core';
import { ToastService } from '../../../core/toast/toast.service';

@Component({
  selector: 'app-toast',
  standalone: true,
  templateUrl: './toast.html',
})
export class ToastContainer {
  toastService = inject(ToastService);
}
