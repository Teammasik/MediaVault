# frontend/app.py
import tkinter as tk
from tkinter import messagebox, filedialog, simpledialog
import os
import requests
from pathlib import Path
from PIL import Image, ImageTk

API_BASE = "http://localhost:8080"
UPLOADS_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "uploads"))
os.makedirs(UPLOADS_DIR, exist_ok=True)


class MediaVaultApp:
    def __init__(self, root):
        self.root = root
        self.root.title("MediaVault — Вход")
        self.root.geometry("1000x700")
        self.root.resizable(False, False)
        self.user_id = None
        self.is_admin = False
        self.files_window = None
        self.show_login()

    def show_login(self):
        self.clear_window()
        self.root.title("MediaVault — Вход")

        tk.Label(self.root, text="Вход", font=("Arial", 16, "bold")).pack(pady=10)

        tk.Label(self.root, text="Имя пользователя:").pack(pady=(10, 0))
        self.login_username = tk.Entry(self.root, width=30)
        self.login_username.pack()

        tk.Label(self.root, text="Пароль:").pack(pady=(10, 0))
        self.login_password = tk.Entry(self.root, width=30, show="*")
        self.login_password.pack()

        self.login_error = tk.Label(self.root, text="", fg="red")
        self.login_error.pack(pady=5)

        tk.Button(self.root, text="Войти", command=self.handle_login, width=20).pack(pady=10)
        tk.Button(self.root, text="Нет аккаунта? Зарегистрироваться", 
                  command=self.show_register, 
                  relief="flat", 
                  fg="blue", 
                  cursor="hand2").pack()

    def show_register(self):
        self.clear_window()
        self.root.title("MediaVault — Регистрация")

        tk.Label(self.root, text="Регистрация", font=("Arial", 16, "bold")).pack(pady=10)

        tk.Label(self.root, text="Имя пользователя:").pack(pady=(10, 0))
        self.reg_username = tk.Entry(self.root, width=30)
        self.reg_username.pack()

        tk.Label(self.root, text="Email:").pack(pady=(10, 0))
        self.reg_email = tk.Entry(self.root, width=30)
        self.reg_email.pack()

        tk.Label(self.root, text="Пароль:").pack(pady=(10, 0))
        self.reg_password = tk.Entry(self.root, width=30, show="*")
        self.reg_password.pack()

        self.reg_error = tk.Label(self.root, text="", fg="red")
        self.reg_error.pack(pady=5)

        tk.Button(self.root, text="Зарегистрироваться", command=self.handle_register, width=20).pack(pady=10)
        tk.Button(self.root, text="Назад к входу", command=self.show_login, relief="flat", fg="blue", cursor="hand2").pack()

    def handle_login(self):
        username = self.login_username.get().strip()
        password = self.login_password.get().strip()

        if not username or not password:
            self.login_error.config(text="Заполните все поля")
            return

        try:
            resp = requests.post(f"{API_BASE}/login", json={
                "username": username,
                "password": password
            }, timeout=5)
            
            if resp.status_code == 200:
                data = resp.json()
                self.user_id = data.get("user_id")
                self.is_admin = data.get("is_admin", False)
                self.show_main()
            else:
                msg = resp.json().get("message", "Неверное имя или пароль")
                self.login_error.config(text=msg)
        except Exception as e:
            self.login_error.config(text=f"Ошибка подключения: {str(e)}")

    def handle_register(self):
        username = self.reg_username.get().strip()
        email = self.reg_email.get().strip()
        password = self.reg_password.get().strip()

        if not username or not email or not password:
            self.reg_error.config(text="Заполните все поля")
            return

        try:
            resp = requests.post(f"{API_BASE}/register", json={
                "username": username,
                "email": email,
                "password": password
            }, timeout=5)

            if resp.status_code == 201:
                messagebox.showinfo("Успех", "Регистрация прошла успешно!")
                self.show_login()
            else:
                msg = resp.json().get("message", "Ошибка регистрации")
                self.reg_error.config(text=msg)
        except Exception as e:
            self.reg_error.config(text=f"Ошибка подключения: {str(e)}")

    def upload_file(self):
        filepath = filedialog.askopenfilename(
            title="Выберите файл",
            filetypes=[
                ("Изображения", "*.jpg *.jpeg *.png"),
                ("Документы", "*.pdf *.doc *.docx"),
                ("Все файлы", "*.*")
            ]
        )
        if not filepath:
            return

        caption = simpledialog.askstring("Подпись", "Введите подпись к файлу:")
        if caption is None:
            return

        try:
            with open(filepath, "rb") as f:
                files = {"file": (os.path.basename(filepath), f, "application/octet-stream")}
                data = {"caption": caption}
                headers = {"X-User-ID": str(self.user_id)}

                resp = requests.post(
                    f"{API_BASE}/upload",
                    headers=headers,
                    files=files,
                    data=data,
                    timeout=10
                )

            if resp.status_code == 201:
                self.upload_result.config(text="Файл загружен!", fg="green")
            else:
                msg = resp.json().get("message", "Ошибка загрузки")
                self.upload_result.config(text=f"{msg}", fg="red")
        except Exception as e:
            self.upload_result.config(text=f"Ошибка: {str(e)}", fg="red")

    def show_files(self):
        if self.files_window is not None and self.files_window.winfo_exists():
            self.files_window.destroy()
        self.files_window = tk.Toplevel(self.root)
        self.files_window.title("Все файлы")
        self.files_window.geometry("600x500")

        canvas = tk.Canvas(self.files_window)
        scrollbar = tk.Scrollbar(self.files_window, orient="vertical", command=canvas.yview)
        scrollable_frame = tk.Frame(canvas)

        scrollable_frame.bind(
            "<Configure>",
            lambda e: canvas.configure(scrollregion=canvas.bbox("all"))
        )
        canvas.create_window((0, 0), window=scrollable_frame, anchor="nw")
        canvas.configure(yscrollcommand=scrollbar.set)

        file_records = []
        try:
            resp = requests.get(f"{API_BASE}/files", timeout=5)
            if resp.status_code == 200:
                file_records = resp.json()
            else:
                tk.Label(scrollable_frame, text="Ошибка загрузки файлов", fg="red").pack(pady=20)
        except Exception as e:
            tk.Label(scrollable_frame, text=f"Ошибка: {e}", fg="red").pack(pady=20)

        if not file_records:
            tk.Label(scrollable_frame, text="Нет файлов", font=("Arial", 12)).pack(pady=20)
        else:
            row_frame = None
            for i, record in enumerate(file_records):
                if i % 3 == 0:
                    row_frame = tk.Frame(scrollable_frame)
                    row_frame.pack(pady=5)

                disk_filename = record["file_path"]
                display_filename = record["filename"]
                caption = record["caption"]
                created_by = record["created_by"]
                owner_id = record["user_id"]
                file_id = record["id"]

                file_path = os.path.join(UPLOADS_DIR, disk_filename)
                self.create_file_card_with_caption(
                    row_frame, 
                    file_path, 
                    display_filename, 
                    caption, 
                    created_by, 
                    owner_id, 
                    file_id
                )

        canvas.pack(side="left", fill="both", expand=True)
        scrollbar.pack(side="right", fill="y")

    def download_file(self, file_id, filename):
        try:
            resp = requests.get(f"{API_BASE}/download?id={file_id}", timeout=10)
            if resp.status_code != 200:
                messagebox.showerror("Ошибка", "Не удалось скачать файл")
                return

            save_path = filedialog.asksaveasfilename(
                initialfile=filename,
                title="Сохранить файл как...",
                defaultextension=Path(filename).suffix,
                filetypes=[("Все файлы", "*.*")]
            )
            if not save_path:
                return

            with open(save_path, "wb") as f:
                f.write(resp.content)

            messagebox.showinfo("Успех", "Файл сохранён!")

        except Exception as e:
            messagebox.showerror("Ошибка", f"Не удалось сохранить: {str(e)}")

    def create_file_card_with_caption(self, parent, file_path, display_filename, caption, created_by, owner_id, file_id):
        card = tk.Frame(parent, width=150, height=240, relief="raised", bd=1)
        card.pack(side="left", padx=10, pady=5)
        card.pack_propagate(False)

        clickable_frame = tk.Frame(card, width=120, height=120)
        clickable_frame.pack(pady=5)
        clickable_frame.pack_propagate(False)

        ext = Path(display_filename).suffix.lower()
        is_image = ext in (".jpg", ".jpeg", ".png")

        if is_image and os.path.exists(file_path):
            try:
                img = Image.open(file_path)
                img.thumbnail((120, 120))
                photo = ImageTk.PhotoImage(img)
                img_btn = tk.Button(
                    clickable_frame,
                    image=photo,
                    command=lambda fid=file_id, name=display_filename: self.download_file(fid, name),
                    borderwidth=0,
                    highlightthickness=0
                )
                img_btn.image = photo
                img_btn.pack(expand=True)
            except Exception:
                icon_btn = tk.Button(
                    clickable_frame,
                    text="📷",
                    font=("Arial", 24),
                    command=lambda fid=file_id, name=display_filename: self.download_file(fid, name)
                )
                icon_btn.pack(expand=True)
        else:
            icon_map = {
                ".pdf": "📄", ".doc": "WORD", ".docx": "WORD",
                ".txt": "📝", ".xlsx": "📊", ".pptx": "📽️",
            }
            icon = icon_map.get(ext, "📄")
            icon_btn = tk.Button(
                clickable_frame,
                text=icon,
                font=("Arial", 24),
                command=lambda fid=file_id, name=display_filename: self.download_file(fid, name)
            )
            icon_btn.pack(expand=True)

        name_to_show = (display_filename[:15] + "...") if len(display_filename) > 15 else display_filename
        tk.Label(card, text=name_to_show, font=("Arial", 8), wraplength=140).pack()

        creator_text = f"👤 {created_by}"
        tk.Label(card, text=creator_text, font=("Arial", 7), fg="blue", wraplength=140).pack(pady=(2, 0))

        if caption:
            cap_to_show = (caption[:25] + "...") if len(caption) > 25 else caption
            tk.Label(card, text=cap_to_show, font=("Arial", 7), fg="gray", wraplength=140).pack(pady=(2, 0))

        if owner_id == self.user_id or self.is_admin:
            del_btn = tk.Button(
                card, 
                text="🗑️ Удалить", 
                command=lambda fid=file_id: self.delete_file(fid), 
                font=("Arial", 6), 
                fg="red"
            )
            del_btn.pack(pady=(2, 0))
            
            edit_btn = tk.Button(
                card,
                text="✏️ Подпись",
                command=lambda fid=file_id, curr=caption: self.edit_caption(fid, curr),
                font=("Arial", 6),
                fg="green"
            )
            edit_btn.pack(pady=(2, 0))

    def edit_caption(self, file_id, current_caption):
        new_caption = simpledialog.askstring(
            "Изменить подпись", 
            "Введите новую подпись:", 
            initialvalue=current_caption
        )
        if new_caption is None:
            return
        try:
            headers = {"X-User-ID": str(self.user_id)}
            data = {"file_id": file_id, "caption": new_caption}
            resp = requests.post(f"{API_BASE}/edit-caption", headers=headers, json=data, timeout=5)
            
            if resp.status_code == 200:
                messagebox.showinfo("Успех", "Подпись обновлена")
                self.show_files()
            else:
                msg = resp.json().get("message", "Ошибка")
                messagebox.showerror("Ошибка", msg)
        except Exception as e:  
            messagebox.showerror("Ошибка", f"Не удалось обновить: {e}")

    def delete_file(self, file_id):
        if messagebox.askyesno("Удаление", "Вы уверены, что хотите удалить файл?"):
            try:
                headers = {"X-User-ID": str(self.user_id)}
                resp = requests.post(f"{API_BASE}/delete-file?id={file_id}", headers=headers, timeout=5)
                if resp.status_code == 200:
                    messagebox.showinfo("Успех", "Файл удалён")
                    self.show_files()
                else:
                    msg = resp.json().get("message", "Ошибка удаления")
                    messagebox.showerror("Ошибка", msg)
            except Exception as e:
                messagebox.showerror("Ошибка", f"Не удалось удалить: {e}")

    def show_main(self):
        self.clear_window()
        self.root.title("MediaVault — Главная")

        user_info = f"ID: {self.user_id}"
        if self.is_admin:
            user_info += " (админ)"
        tk.Label(self.root, text="Добро пожаловать!", font=("Arial", 16, "bold")).pack(pady=10)
        tk.Label(self.root, text=user_info, font=("Arial", 10)).pack()

        tk.Button(self.root, text="Загрузить файл", command=self.upload_file, width=20).pack(pady=5)
        tk.Button(self.root, text="К файлам", command=self.show_files, width=20).pack(pady=5)
        tk.Button(self.root, text="Выйти", command=self.show_login, width=20).pack(pady=10)

        self.upload_result = tk.Label(self.root, text="", fg="green")
        self.upload_result.pack(pady=5)

    def clear_window(self):
        for widget in self.root.winfo_children():
            widget.destroy()


if __name__ == "__main__":
    root = tk.Tk()
    app = MediaVaultApp(root)
    root.mainloop()
