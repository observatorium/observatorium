#!/bin/bash

set -e

export GOROOT=$(go env GOROOT)

#PREV="0.0.1"
LATEST="0.0.1"

if [[ -z "$CSV_VERSION" ]]; then
  CSV_VERSION=$LATEST
  REPLACES_CSV_VERSION=$PREV
fi

PACKAGE_NAME="observatorium-operator"

PACKAGE_DIR="${CATALOG_DIR}/${PACKAGE_NAME}"
CSV_DIR="${PACKAGE_DIR}/${CSV_VERSION}"
CSV_FILE="$CSV_DIR/${PACKAGE_NAME}.v${CSV_VERSION}.clusterserviceversion.yaml"

OUT_DIR="_output/olm-catalog"
OUT_CSV_DIR="${OUT_DIR}/${PACKAGE_NAME}/${CSV_VERSION}"
OUT_CSV_FILE="${OUT_CSV_DIR}/${PACKAGE_NAME}.v${CSV_VERSION}.clusterserviceversion.yaml"

ANNOTATIONS_FILE="${DEPLOY_DIR}/csv/annotations.json"
MAINTAINERS_FILE="${DEPLOY_DIR}/csv/maintainers.json"

EXTRA_ANNOTATIONS=""
MAINTAINERS=""

if [ -n "$DESCRIPTION_FILE" ]; then
	DESCRIPTION="-description-from=$DESCRIPTION_FILE"
fi
if [ -n "$MAINTAINERS_FILE" ]; then
	MAINTAINERS="-maintainers-from=$MAINTAINERS_FILE"
fi
if [ -n "$ANNOTATIONS_FILE" ]; then
	EXTRA_ANNOTATIONS="-annotations-from=$ANNOTATIONS_FILE"
fi

SEMVER_REGEX="^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
if ! [[ "$CSV_VERSION" =~ $SEMVER_REGEX ]]; then
	echo "CSV_VERSION not provided or does not match semver format"
	exit 1
fi

# clean up all old data first
rm -rf "$OUT_DIR"
mkdir -p "$OUT_CSV_DIR"

# do not generate new CRD/CSV for old versions
if [[ "$CSV_VERSION" != "$PREV" ]]; then
#  $OPERATOR_SDK generate crds

  # generate a temporary csv we'll use as a template
  $OPERATOR_SDK generate csv \
    --operator-name="${PACKAGE_NAME}" \
    --csv-version="${CSV_VERSION}" \
    --csv-channel="${CSV_VERSION}" \
    --default-channel=true \
    --make-manifests=false \
    --crd-dir=deploy/csv/crds \
    --update-crds

  # using the generated CSV, create the real CSV by injecting all the right data into it
  $GENERATOR \
    --csv-version "${CSV_VERSION}" \
    --operator-csv-template-file "${CSV_FILE}" \
    --operator-image "${FULL_OPERATOR_IMAGE}" \
    --olm-bundle-directory "$OUT_CSV_DIR" \
    --replaces-csv-version "$REPLACES_CSV_VERSION" \
    --skip-range "$CSV_SKIP_RANGE" \
    "${MAINTAINERS}" \
    "${EXTRA_ANNOTATIONS}"
fi

# copy remaining manifests to final location
cp -a $OUT_DIR/ $CATALOG_DIR/../

echo "New OLM manifests created at $CATALOG_DIR"
