import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

interface Editor {
  name: string;
  id: string;
}

@Component({
  selector: 'app-documentation',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './docs.html',
  styleUrls: ['./docs.css']
})
export class DocumentationComponent {

  selectedEditorId: string = 'vs';

  editors: Editor[] = [

    { name: 'VS Code', id: 'vs' },
    { name: 'IntelliJ IDEA', id: 'ij' },
    { name: 'PyCharm', id: 'pc' },
    { name: 'WebStorm', id: 'ws' },
    { name: 'GoLand', id: 'go' },

    { name: 'Rider', id: 'rd' },
    { name: 'CLion', id: 'cl' },
    { name: 'RubyMine', id: 'rm' },
    { name: 'PhpStorm', id: 'ps' },
    { name: 'DataGrip', id: 'dg' },

    { name: 'Android Studio', id: 'android' },
    { name: 'RustRover', id: 'rr' },
    { name: 'Neovim', id: 'neovim' }
  ];

  // دالة لاختيار المحرر وتفعيل الإطار البرتقالي حوله
  selectEditor(id: string) {
    this.selectedEditorId = id;
  }
}