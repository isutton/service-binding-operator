Feature: Resource Scoped Annotations
    @disabled
    Scenario: Copy a single key as string from the Service resource itself to the binding secret
        Given CRD "databases.postgresql.baiju.dev" contains the annotation "service.binding/uri: path={.status.data.url}"
        And Resource "rsa-1-service" is created
            """
            apiVersion: postgresql.baiju.dev/v1alpha1
            kind: Database
            metadata:
                name: rsa-1-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                bootstrap:
                    - type: plain
                      url: myhost2.example.com
                      name: hostGroup1
                    - type: tls
                      url: myhost1.example.com:9092,myhost2.example.com:9092
                      name: hostGroup2
                data:
                    dbConfiguration: database-config     # ConfigMap
                    dbCredentials: database-cred-Secret  # Secret
                    url: db.stage.ibm.com
            """
        When Resource "rsa-1" is created
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: annotations-1
            spec:
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: rsa-1-service
            """
        Then Secret "rsa-1" contains "URI" key with value "db.stage.ibm.com"

    @disabled
    Scenario: Copy a single key as a map to the Service resource itself to the binding secret
        Given CRD "databases.postgresql.baiju.dev" contains the annotation "service.binding/spec: path={.spec}"
        And Resource "rsa-2-service" is created
            """
            apiVersion: postgresql.baiju.dev/v1alpha1
            kind: Database
            metadata:
                name: rsa-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                bootstrap:
                    - type: plain
                      url: myhost2.example.com
                      name: hostGroup1
                    - type: tls
                      url: myhost1.example.com:9092,myhost2.example.com:9092
                      name: hostGroup2
                data:
                    dbConfiguration: database-config     # ConfigMap
                    dbCredentials: database-cred-Secret  # Secret
                    url: db.stage.ibm.com
            """
        When Resource "rsa-2" is created
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: rsa-1
            spec:
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: rsa-2-service
            """
        Then Secret "rsa-2" contains "SPEC_IMAGE" key with value "docker.io/postgres"
        And Secret "rsa-2" contains "SPEC_IMAGENAME" key with value "postgres"
        And Secret "rsa-2" contains "SPEC_DBNAME" key with value "db-demo"
