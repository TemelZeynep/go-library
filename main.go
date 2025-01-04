package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var books []Book

func main() {
	loadBooksFromFile("db.json") // JSON dosyasından verileri yükle.
	fmt.Println("Server is running on http://localhost:3000")
	http.HandleFunc("/books", booksHandler)
	http.HandleFunc("/books/", bookHandler)
	http.ListenAndServe(":3000", nil)

}

func loadBooksFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var data struct {
		Books []Book `json:"books"`
	}

	if err := json.Unmarshal(byteValue, &data); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	books = data.Books
	fmt.Println("Books loaded successfully:", books)
}

// GET/books
func booksHandler(w http.ResponseWriter, r *http.Request) {
	//Gelen HTTP isteğinin (request) metodunu temsil eder.
	//GET, POST, PUT, DELETE gibi HTTP metodlarını içerir.
	if r.Method == http.MethodGet {
		//"Content-Type" adlı başlığı "application/json" değerine ayarlıyorum.
		w.Header().Set("Content-Type", "application/json")
		//Gelen veriyi JSON formatına dönüştürüyorum.
		json.NewEncoder(w).Encode(books)
		return
	} else if r.Method == http.MethodPost {
		addBook(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "Failed to encode books", http.StatusInternalServerError)
	}
}

// POST/books
func addBook(w http.ResponseWriter, r *http.Request) {
	// Book structına ait newBook nesnesi oluşturuyorum.
	var newBook Book

	//HTTP isteğinin gövdesini temsil eder.
	//İstemciden gelen JSON formatındaki yeni kitap verisini çözümler.
	if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		fmt.Println("Error decoding JSON:", err) // Hata ayıklama için logla
		return
	}
	// Bu newBook'a ID atıyorum.
	newBook.ID = len(books) + 1          //ID atıyorum
	books = append(books, newBook)       //Yeni kitabı mevcut kitap listesine ekliyorum.
	fmt.Println("Current books:", books) // Slice içeriğini terminalde kontrol edin

	//HTTP durum kodu 201dir. Bu kod, istemciye
	//"Kaynak başarıyla oluşturuldu" anlamında bir mesaj iletiyorum.

	// Dosyayı güncelle.
	saveBooksToFile("db.json")
	w.WriteHeader(http.StatusCreated)
	//newBook nesnesini JSON formatına dönüştürüyorum
	if err := json.NewEncoder(w).Encode(newBook); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		fmt.Println("Error encoding JSON:", err) // Hata ayıklama için logla
		return
	}
}

func saveBooksToFile(s string) {
	panic("unimplemented")
}

// PUT/books/{id}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	//bir HTTP isteğinden gelen URL'deki kitap ID'si alıyorum.
	//Bir tam sayı (integer) değerine dönüştürüyorum.
	idStr := strings.TrimPrefix(r.URL.Path, "/books/") //İstekteki URL'den /books/ öneki kaldırılarak yalnızca ID kısmı alıyorum.
	id, err := strconv.Atoi(idStr)                     // string olarak alınan id'yi integer'a çeviyorum.
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid book ID: %s", idStr), http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		updateBook(w, r, id)
	case http.MethodDelete:
		deleteBook(w, r, id)
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

}

func updateBook(w http.ResponseWriter, r *http.Request, id int) {
	var updatedBook Book
	json.NewDecoder(r.Body).Decode(&updatedBook)
	for i, book := range books { // Books slice'da sırayla dolanır.
		// "book": her dolanışta slice'daki kitabın kopyasını temsil eder.
		if book.ID == id { // İlgili kitabın id'si aranan kitabın id'si mi?
			books[i] = updatedBook // Eğer öyleyse ilgili kitaba güncel bilgileri ata.
			books[i].ID = id       //Atama sonrası bir kontrol daha.
			json.NewEncoder(w).Encode(books[i])
			// ".NewEncoder(w)" ile Http yanıtını JSOn formatında kodlamak için. Bu işlem ile istemci kitabı JSON formatında görüntüler.
			// Güncellenmiş kitap nesnesi JSON formatında gönderilir.
			return
		}

	}
	http.Error(w, "Book not found", http.StatusNotFound)

}

func deleteBook(w http.ResponseWriter, r *http.Request, id int) {
	for i, book := range books { // Books slice'da sırayla dolanır.
		// "book": her dolanışta slice'daki kitabın kopyasını temsil eder.
		if book.ID == id {
			books = append(books[:i], books[i+1:]...)
			// İlgili indekse sahip kitaptan önceki ve sonra kitapları birleştir. Yeni yapıyı booksa ata
			w.WriteHeader(http.StatusNoContent) // Kullanıcıya 204 dön.
			return

		}

	}
	http.Error(w, "Book not found", http.StatusNotFound)
}
