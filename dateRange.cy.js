describe('Date Range Functionality Test', () => {
  beforeEach(() => {
    // Assuming your UI server is running on localhost:3000
    cy.visit('http://localhost:3000');
  });

  const selectDateRange = (startDate, endDate) => {
    cy.get('#startDateInput').clear().type(startDate);
    cy.get('#endDateInput').clear().type(endDate);
    cy.get('#applyButton').click();
  };

  const checkResult = (expectedResult) => {
    cy.get('#result').should('contain', expectedResult);
  };

  const checkFetchedDates = () => {
    cy.get('#fetchDatesButton').click();

    cy.get('#startDate').invoke('text').then((startDate) => {
      expect(startDate).to.match(/^\d{4}-\d{2}-\d{2}$/);
    });

    cy.get('#endDate').invoke('text').then((endDate) => {
      expect(endDate).to.match(/^\d{4}-\d{2}-\d{2}$/);
    });
  };

  it('Selects a Valid Date Range and Verifies Result', () => {
    selectDateRange('2024-03-01', '2024-03-10');
    checkResult('Selected Date Range: 2024-03-01 - 2024-03-10');
  });

  it('Displays Error for Invalid Date Range', () => {
    selectDateRange('2024-03-10', '2024-03-01');
    checkResult('Error: Invalid Date Range');
  });

  it('Checks Manually Typed Date Range', () => {
    selectDateRange('2024-03-05', '2024-03-15');
    checkResult('Selected Date Range: 2024-03-05 - 2024-03-15');
  });

  it('Selects Last 7 Days', () => {
    // Logic to calculate last 7 days start and end dates
    const last7DaysStartDate = '...';
    const last7DaysEndDate = '...';

    selectDateRange(last7DaysStartDate, last7DaysEndDate);
    checkResult(`Selected Date Range: ${last7DaysStartDate} - ${last7DaysEndDate}`);
  });

  it('Selects Last 30 Days', () => {
    // Logic to calculate last 30 days start and end dates
    const last30DaysStartDate = '...';
    const last30DaysEndDate = '...';

    selectDateRange(last30DaysStartDate, last30DaysEndDate);
    checkResult(`Selected Date Range: ${last30DaysStartDate} - ${last30DaysEndDate}`);
  });

  it('Selects Last 365 Days', () => {
    // Logic to calculate last 365 days start and end dates
    const last365DaysStartDate = '...';
    const last365DaysEndDate = '...';

    selectDateRange(last365DaysStartDate, last365DaysEndDate);
    checkResult(`Selected Date Range: ${last365DaysStartDate} - ${last365DaysEndDate}`);
  });

  it('Selects Next 365 Days', () => {
    // Logic to calculate next 365 days start and end dates
    const next365DaysStartDate = '...';
    const next365DaysEndDate = '...';

    selectDateRange(next365DaysStartDate, next365DaysEndDate);
    checkResult(`Selected Date Range: ${next365DaysStartDate} - ${next365DaysEndDate}`);
  });

  it('Clears Both Start and End Date', () => {
    selectDateRange('2024-03-01', '2024-03-10');
    cy.get('#clearButton').click();
    cy.get('#startDateInput').should('have.value', '');
    cy.get('#endDateInput').should('have.value', '');
  });

  it('Fetches Dates from UI Server', () => {
    checkFetchedDates();
  });
});
