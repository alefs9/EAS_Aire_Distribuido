package persistencia

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestCacheMissHitExpiracionYPersistencia(t *testing.T) {
	s, ctx := servicioPrueba(t)
	id, result := crearFixture(t, s, ctx)
	for _, estado := range []string{"queued", "running"} {
		j, err := s.Resultado(ctx, id)
		if !errors.Is(err, ErrPendiente) || j.Trabajo.Estado != estado {
			t.Fatal(j, err)
		}
		if n, _ := s.Cache.Exists(ctx, s.Clave(id)).Result(); n != 0 {
			t.Fatal("trabajo pendiente cacheado")
		}
		if estado == "queued" {
			if err = s.Iniciar(ctx, id); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := s.Completar(ctx, id, result); err != nil {
		t.Fatal(err)
	}
	a, err := s.Resultado(ctx, id)
	if err != nil || a.Fuente != "MongoDB" {
		t.Fatal(a, err)
	}
	b, err := s.Resultado(ctx, id)
	if err != nil || b.Fuente != "Redis" || !reflect.DeepEqual(a.Trabajo, b.Trabajo) {
		t.Fatal(b, err)
	}
	ttl, err := s.Cache.PTTL(ctx, s.Clave(id)).Result()
	if err != nil || ttl <= 0 || ttl > s.TTL {
		t.Fatal(ttl, err)
	}
	time.Sleep(s.TTL + 100*time.Millisecond)
	c, err := s.Resultado(ctx, id)
	if err != nil || c.Fuente != "MongoDB" {
		t.Fatal(c, err)
	}
	if err = s.Invalidar(ctx, id); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.Cache.Exists(ctx, s.Clave(id)).Result(); n != 0 {
		t.Fatal("clave no invalidada")
	}
	d, err := s.Resultado(ctx, id)
	if err != nil || d.Fuente != "MongoDB" || !reflect.DeepEqual(d.Trabajo, a.Trabajo) {
		t.Fatal("resultado perdido al vaciar caché", err)
	}
}

func TestRedisNoDisponible(t *testing.T) {
	s, ctx := servicioPrueba(t)
	id, r := crearFixture(t, s, ctx)
	if err := s.Iniciar(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := s.Completar(ctx, id, r); err != nil {
		t.Fatal(err)
	}
	// Puerto reservado sin servicio: no detiene ni modifica el Redis compartido.
	c := NuevaCache("127.0.0.1:1", "")
	defer c.Close()
	s.Cache = c
	x, err := s.Resultado(ctx, id)
	if err != nil || x.Fuente != "MongoDB" || x.Aviso == "" {
		t.Fatal(x, err)
	}
}
