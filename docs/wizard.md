# Asistente de configuración

Este primer hito permite seleccionar un establecimiento del registro DEIS,
definir su organización, guardar y confirmar un plan, y descargarlo como JSON.
Utiliza los colores, fuentes y logotipos existentes de Gestión.

**Todavía no instala AIO, aplica el tema a Nextcloud, crea usuarios ni instala
aplicaciones.** Confirmar el plan guarda una configuración; no aplica cambios a
la suite. No es una versión lista para producción.

## Ejecutar una compilación local

Requiere la versión de Go indicada en [go.mod](../go.mod). Desde el repositorio:

```sh
go build -o bin/modding ./cmd/modding
./bin/modding --state-dir .modding
```

Abra `http://127.0.0.1:8787` en el mismo equipo. Copie el código de acceso que
aparece en la terminal. El código funciona una sola vez y la sesión dura hasta
ocho horas. Al reiniciar el proceso se invalidan las sesiones y se genera un
código nuevo; la configuración guardada permanece. Este mecanismo de desarrollo
no sustituye el futuro inicio de sesión permanente del administrador.

El proceso no necesita root ni acceso a Docker. Utilice un directorio exclusivo
para el estado. Se crea con permisos `0700` y la base SQLite con `0600`; se rechazan
directorios compartidos y archivos de estado que sean enlaces simbólicos.

Para abrir el asistente en una IP concreta de la red, proporcione un certificado
y su clave, válidos para esa IP y confiables para los equipos que se conectan:

```sh
./bin/modding --listen 192.168.1.20:8787 --state-dir .modding \
  --tls-cert /ruta/al/certificado.pem --tls-key /ruta/a/la/clave.pem
```

Sustituya la IP y las rutas por los valores de su entorno. Se rechaza una escucha
fuera de loopback sin TLS, y también las direcciones comodín. Abra exactamente
la IP y el puerto indicados por el proceso. Esta versión atiende conexiones
directas; no interpreta encabezados de un proxy. Limite el acceso a la LAN/VPN.
La provisión automática del certificado de arranque sigue pendiente.

## Elegir el establecimiento

1. Busque por nombre, código DEIS, comuna o dirección. La búsqueda ignora
   mayúsculas y acentos. Puede limitarla por región.
2. Revise el nombre, la comuna, la dirección y el código del resultado.
3. Pulse **Seleccionar**. Ningún establecimiento viene seleccionado por defecto.

El registro integrado contiene 2.655 establecimientos, con fecha **2026-07-23**.
Es el archivo utilizado por el selector de Gestión en la revisión importada,
no el archivo maestro posterior que tiene otro esquema. Consulte la
[procedencia de los datos y recursos](../third_party/gestion/README.md).

Si falta un establecimiento, solicite una actualización del registro. No se
admiten identidades DEIS inventadas ni un reemplazo por entrada libre.

## Configurar y revisar

Ingrese el nombre de dominio de la suite y los sectores, programas y unidades
que correspondan, un nombre por línea. Las listas pueden quedar vacías. No
incluya información de pacientes ni datos personales en los nombres de equipos.

La vista previa propone un grupo por equipo y sus carpetas. El área Transversal
es de lectura para `all-staff`; cada carpeta de equipo propone lectura y escritura
sin compartir para su grupo. Son propuestas de este primer hito, no permisos ya
aplicados ni una importación completa de todas las reglas de Gestión. La asignación
de usuarios y la edición detallada de permisos se implementarán con el aprovisionamiento.

Estadística, Farmacia y Tótem comienzan desactivadas. Puede incluirlas en el plan.
Las demás aplicaciones del catálogo acordado aparecen como incluidas, con su
fuente prevista. La compatibilidad e instalación de esas aplicaciones siguen
pendientes; la lista no certifica que ya estén instaladas.

Pulse **Guardar y revisar**, revise la identidad y el acceso propuesto, y pulse
**Confirmar plan**. Puede descargar el JSON o editarlo nuevamente. Editar devuelve
el plan al estado de borrador. Elegir otro establecimiento inicia una configuración
nueva para evitar conservar sectores o aplicaciones de la selección anterior.

Dos pestañas no pueden sobrescribir silenciosamente la misma versión: si otra
pestaña guardó primero, se solicita recargar y revisar la configuración actual.

## Cambiar el catálogo

El ejecutable admite un CSV externo con las mismas columnas que el catálogo
integrado, incluida la identidad territorial y las coordenadas:

```sh
./bin/modding --state-dir .modding \
  --catalog /ruta/establecimientos-deis.csv --catalog-version FECHA-DEL-REGISTRO
```

Use la fecha o versión real de la fuente. Se rechazan archivos vacíos, columnas
obligatorias ausentes, códigos inválidos y códigos repetidos. Al cargar un archivo
distinto, su SHA-256 permite detectar cambios incluso si conserva el mismo nombre
de versión.

La configuración anterior conserva su identidad y la versión que aprobó. Para
continuar, vuelva a seleccionar el mismo código y revise el registro actualizado.
Esto conserva las opciones de organización pero vuelve a dejar el plan en borrador.
Si el código desapareció, debe resolver la actualización del registro antes de
confirmar. El asistente no lo sustituye por otro establecimiento.

## Verificación para desarrollo

```sh
go test -race ./...
go vet ./...
python3 scripts/check_sources.py
python3 scripts/check_docs.py
```

Para ejecutar el recorrido de navegador con establecimientos ficticios:

```sh
python3 -m venv /tmp/modding-browser-env
/tmp/modding-browser-env/bin/pip install -r tests/requirements.txt
/tmp/modding-browser-env/bin/python -m playwright install chromium
/tmp/modding-browser-env/bin/python tests/browser_smoke.py --binary ./bin/modding
```

El recorrido verifica selección, configuración, exportación, reinicio con TLS,
cookies protegidas, cambio de establecimiento y presentación móvil. Usa un
certificado temporal solo para la prueba y nunca modifica una instalación AIO.
Las comprobaciones de integración completa y recuperación de la suite permanecen
en el [plan de validación](validation.md).
