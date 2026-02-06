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

	// 1. OTK Template (ОТК) - ТЕПЕРЬ ТОЖЕ ОБНОВЛЯЕМ (удаляем старый)
	otkQuestions := []domain.Question{
		{
			Text:            "Все полки сзади закреплены стяжками для фиксации при грузоперевозке, чтобы минимизировать выпадение контактов электрики из цепи с дисплеями",
			Order:           1,
			MinPhotos:       1,
			MaxPhotos:       5,
			IsRequired:      true,
			ReferenceImages: []string{"refs/otk_example_1.jpg"}, // Путь к фото в B2
		},
		{
			Text:            "Пружины в каждой ячейке выставлены таким образом, чтобы из них не вываливался SOKOLOV сюрпрайз.",
			Order:           2,
			MinPhotos:       1,
			MaxPhotos:       5,
			IsRequired:      true,
			ReferenceImages: []string{"refs/otk_example_2.jpg"},
		},
		{
			Text:            "Дисплеи всех линий горят, сняты светофильтры для того, чтобы дисплей горел ярче и лучше светилось табло.",
			Order:           3,
			MinPhotos:       1,
			MaxPhotos:       5,
			IsRequired:      true,
			ReferenceImages: []string{"refs/otk_example_3.jpg"},
		},
		{Text: "Установлена сим-карта, и рядом на двери наклеена сопроводительная информация с номером сим-карты и номером телефона из комплекта SOKOLOV.", Order: 4, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
		{Text: "На месте монетоприемника стоит металлическая заглушка", Order: 5, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleOTK, otkQuestions, true)

	// 2. Sticker Template (Оклейка) - Удаляем старый перед созданием
	stickerQuestions := []domain.Question{
		{Text: "Тестовый пункт оклейки: Проверка качества нанесения фронтальной пленки", Order: 1, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
		{Text: "Тестовый пункт оклейки: Отсутствие воздушных пузырей и заломов", Order: 2, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleSticker, stickerQuestions, true)

	// 3. Ads Template (Реклама) - Удаляем старый перед созданием
	adsQuestions := []domain.Question{
		{Text: "Тестовый пункт рекламы: Лайтбокс установлен и надежно закреплен", Order: 1, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
		{Text: "Тестовый пункт рекламы: Проверка равномерности подсветки", Order: 2, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleAds, adsQuestions, true)

	// 4. Assembler Template (Сборка) - Удаляем старый перед созданием
	assemblerQuestions := []domain.Question{
		{Text: "Тестовый пункт сборки: Проверка затяжки всех силовых контактов", Order: 1, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
		{Text: "Тестовый пункт сборки: Укладка кабелей внутри корпуса", Order: 2, MinPhotos: 1, MaxPhotos: 5, IsRequired: true},
	}
	seedTemplate(ctx, templateUC, domain.RoleAssembler, assemblerQuestions, true)

	log.Println("Seeding completed successfully!")
}

func seedTemplate(ctx context.Context, uc *usecase.TemplateUseCase, role domain.Role, questions []domain.Question, overwrite bool) {
	if overwrite {
		log.Printf("Deleting old template for role %s...\n", role)
		_ = uc.DeleteTemplateByRole(ctx, role)
	} else {
		// Check if template exists
		oldTmpl, _, err := uc.GetTemplateByRole(ctx, role)
		if err == nil && oldTmpl != nil {
			log.Printf("Template for role %s already exists, skipping.\n", role)
			return
		}
	}

	_, err := uc.CreateTemplate(ctx, role, questions)
	if err != nil {
		log.Printf("Failed to seed template for %s: %v\n", role, err)
	} else {
		log.Printf("Seeded template for %s\n", role)
	}
}
