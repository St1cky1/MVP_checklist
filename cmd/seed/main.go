package main

import (
	"context"
	"log"
	"os"

	"MVP_checklist/internal/domain"
	"MVP_checklist/internal/repository"
	"MVP_checklist/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/checklist_db"
	}
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	repo := repository.NewPostgresRepository(dbPool)
	templateUC := usecase.NewTemplateUseCase(repo)

	// 1. OTK Template
	otkQuestions := []domain.Question{
		{Text: "Проверка отсутствия царапин на корпусе", Order: 1, MinPhotos: 1, MaxPhotos: 3, IsRequired: true},
		{Text: "Проверка работоспособности купюроприемника", Order: 2, MinPhotos: 1, MaxPhotos: 2, IsRequired: true},
		{Text: "Проверка комплектации (ключи, паспорт)", Order: 3, MinPhotos: 1, MaxPhotos: 1, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleOTK, otkQuestions)

	// 2. Sticker Template
	stickerQuestions := []domain.Question{
		{Text: "Оклейка фронтальной панели", Order: 1, MinPhotos: 1, MaxPhotos: 2, IsRequired: true},
		{Text: "Оклейка боковых панелей", Order: 2, MinPhotos: 2, MaxPhotos: 4, IsRequired: true},
		{Text: "Отсутствие пузырей под пленкой", Order: 3, MinPhotos: 1, MaxPhotos: 3, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleSticker, stickerQuestions)

	// 3. Ads Template
	adsQuestions := []domain.Question{
		{Text: "Установка рекламного лайтбокса", Order: 1, MinPhotos: 1, MaxPhotos: 1, IsRequired: true},
		{Text: "Проверка подсветки", Order: 2, MinPhotos: 1, MaxPhotos: 1, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleAds, adsQuestions)

	// 4. Assembler Template
	assemblerQuestions := []domain.Question{
		{Text: "Монтаж платежной системы", Order: 1, MinPhotos: 1, MaxPhotos: 2, IsRequired: true},
		{Text: "Подключение шлейфов", Order: 2, MinPhotos: 2, MaxPhotos: 3, IsRequired: true},
		{Text: "Настройка модема", Order: 3, MinPhotos: 1, MaxPhotos: 1, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleAssembler, assemblerQuestions)

	log.Println("Seeding completed successfully!")
}

func seedTemplate(ctx context.Context, uc *usecase.TemplateUseCase, role domain.Role, questions []domain.Question) {
	_, _, err := uc.GetTemplateByRole(ctx, role)
	if err == nil {
		log.Printf("Template for role %s already exists, skipping seed.\n", role)
		return
	}

	_, err = uc.CreateTemplate(ctx, role, questions)
	if err != nil {
		log.Printf("Failed to seed template for %s: %v\n", role, err)
	} else {
		log.Printf("Seeded template for %s\n", role)
	}
}
