package model

// Author holds the blog owner's public profile information.
// Update these values to match your current info.
var Author = AuthorProfile{
	Name:        "Mateus Henrique",
	Role:        "Back-end Developer",
	Company:     "Uticket",
	Location:    "Brasil",
	Bio:         "Back-end developer especializado em Go, com experiência em Node.js, Python e Java. Apaixonado por arquitetura de sistemas, qualidade de código e tudo que envolve inteligência artificial. Aqui documento meu caminho nos estudos sobre IA.",
	AvatarURL:   "", // coloque a URL de uma foto se quiser
	LinkedInURL: "https://www.linkedin.com/in/mateus-henrique-da-silva/",
	GitHubURL:   "https://github.com/mateus-henrique-silva",
	Skills: []string{
		"Go", "Node.js", "Python", "Java",
		"PostgreSQL", "MongoDB",
		"Docker", "Kubernetes", "AWS",
	},
}

type AuthorProfile struct {
	Name        string
	Role        string
	Company     string
	Location    string
	Bio         string
	AvatarURL   string
	LinkedInURL string
	GitHubURL   string
	Skills      []string
}
