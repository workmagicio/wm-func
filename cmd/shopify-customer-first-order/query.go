package main

var base_query = `
        query GetUpdatedCustomers {
          customers(%s) {
            pageInfo {
              hasNextPage
              endCursor
            }
            edges {
              cursor
              node {
                id
                email
                numberOfOrders
                firstOrder: orders(first: 1, sortKey: PROCESSED_AT) {
                  edges {
                    node {
                      id
                      processedAt
                    }
                  }
                }
              }
            }
          }
        }`
