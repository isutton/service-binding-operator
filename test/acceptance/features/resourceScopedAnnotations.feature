Feature: Resource Scoped Annotations

    As a user of Service Binding Operator

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Copy a single key as string from the Service resource itself to the binding secret
        Given OLM Operator "backend" is running
        And The Custom Resource Definition is present
        """
        apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
            name: backends.stable.example.com
            annotations:
                service.binding/uri: path={.status.data.url}
        spec:
            group: stable.example.com
            versions:
              - name: v1
                served: true
                storage: true
            scope: Namespaced
            names:
                plural: backends
                singular: backend
                kind: Backend
                shortNames:
                  - bk
        """
        And The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
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
        When Service Binding is applied
        """
        apiVersion: operators.coreos.com/v1alpha1
        kind: ServiceBinding
        metadata:
            name: rsa-1
        spec:
            services:
              - group: stable.example.com
                version: v1
                kind: Backend
                name: rsa-1-service
        """
        Then Secret "rsa-1" contains "BACKEND_URI" key with value "db.stage.ibm.com"

    Scenario: Copy a single key as a map to the Service resource itself to the binding secret
        Given OLM Operator "backend" is running
        And The Custom Resource Definition is present
        """
        apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
            name: backends.stable.example.com
            annotations:
                service.binding/spec: path={.spec},elementType=map
        spec:
            group: stable.example.com
            versions:
              - name: v1
                served: true
                storage: true
            scope: Namespaced
            names:
                plural: backends
                singular: backend
                kind: Backend
                shortNames:
                  - bk
        """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
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
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: rsa-2
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: rsa-2-service
            """
        Then Secret "rsa-2" contains "BACKEND_SPEC_IMAGE" key with value "docker.io/postgres"
        And Secret "rsa-2" contains "BACKEND_SPEC_IMAGENAME" key with value "postgres"
        And Secret "rsa-2" contains "BACKEND_SPEC_DBNAME" key with value "db-demo"
