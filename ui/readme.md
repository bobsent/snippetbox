The ui directory will contain the user-interface assets used by the web application. 
Specifically, the ui/html directory will contain HTML templates, and the ui/static directory 
will contain static files (like CSS and images).


 The .tmpl extension doesn’t convey any special meaning or behavior here. I’ve only chosen this extension because it’s a nice way of making it clear that the file contains a Go template when you’re browsing a list of files. But, if you want, you could use the extension .html instead (which may make your text editor recognize the file as HTML for the purpose of syntax highlighting or autocompletion) — or you could even use a ‘double extension’ like .tmpl.html. The choice is yours, but we’ll stick to using .tmpl for our templates throughout this book.