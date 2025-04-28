package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id, "expected non-zero id after insert")
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, id, storedParcel.Number, "parcel number should match inserted id")
	assert.Equal(t, parcel.Client, storedParcel.Client)
	assert.Equal(t, parcel.Address, storedParcel.Address)
	assert.Equal(t, parcel.Status, storedParcel.Status)
	assert.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err)
	_, err = store.Get(id)
	require.Error(t, err, "expected error when getting deleted parcel")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id, "expected non-zero id after insert")

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, newAddress, storedParcel.Address, "address should be updated")
	assert.Equal(t, parcel.Client, storedParcel.Client)
	assert.Equal(t, parcel.Status, storedParcel.Status)
	assert.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id, "expected non-zero id after insert")

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, newStatus, storedParcel.Status, "status should be updated")
	assert.Equal(t, parcel.Client, storedParcel.Client)
	assert.Equal(t, parcel.Address, storedParcel.Address)
	assert.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	assert.Len(t, storedParcels, len(parcels), "number of parcels should match")

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		original, ok := parcelMap[parcel.Number]
		assert.True(t, ok, "parcel should exist in added map")
		assert.Equal(t, original, parcel)	
	}
}
