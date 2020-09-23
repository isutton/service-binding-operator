Feature: Secret Scoped Annotations
    @disabled
    Scenario: Copy a single key from the related Secret into the binding secret
        Given CRD "databases.postgresql.baiju.dev" contains the annotation "service.binding/username: path={.status.data.dbConfiguration},objectType=Secret,sourceKey=username"
        And Resource "ssa-1-secret" is created
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: ssa-1-secret
            data:
                username: AzureDiamond
            """
        And Resource "ssa-1-service" is created
            """
            apiVersion: postgresql.baiju.dev/v1alpha1
            kind: Database
            metadata:
                name: ssa-1-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssa-1-secret
            """
        When Resource "ssa-1" is created
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: ssa-1
            spec:
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: ssa-1-service
            """
        Then Secret "ssa-1" contains "USERNAME" key with value "AzureDiamond"

    @disabled
    Scenario: Copy all keys from the Secret related to the Service resource to the binding secret
        Given CRD "databases.postgresql.baiju.dev" contains the annotation "service.binding: path={.status.data.dbConfiguration},objectType=Secret,elementType=map"
        And Resource "ssa-2-secret" is created
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: ssa-2-secret
            data:
                username: AzureDiamond
                password: hunter2
            """
        And Resource "ssa-2-service" is created
            """
            apiVersion: postgresql.baiju.dev/v1alpha1
            kind: Database
            metadata:
                name: ssa-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssa-2-secret
            """
        When Resource "ssa-2" is created
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: ssa-2
            spec:
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: ssa-2-service
            """
        Then Secret "ssa-2" contains "USERNAME" key with value "AzureDiamond"
        And Secret "ssa-2" contains "PASSWORD" key with value "hunter2"
